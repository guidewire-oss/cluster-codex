package k8

import (
	"cluster-codex/internal/model"
	"cluster-codex/internal/utils"
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/set"
	"net/url"
	"os"
	"strings"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . K8sClientInterface
type K8sClientInterface interface {
	GetAllComponents(ctx context.Context) ([]model.Component, []string, error)
	GetAllImages(ctx context.Context, namespaceList []string) ([]model.Component, error)
}

// K8sClient is the concrete implementation of the K8sClientInterface
type K8sClient struct {
	K8sContext    string
	Config        *rest.Config
	Client        kubernetes.Interface
	DynamicClient dynamic.Interface
	Discovery     discovery.DiscoveryInterface
}

// pod.Status.EphemeralContainerStatuses has a different return type from
// ContainerStatuses and InitContainerStatuses so use an interface
type ContainerLike interface {
	GetName() string
	GetImage() string
}

// Wrapper for corev1.Container
type ContainerWrapper struct {
	corev1.Container
}

var K8Filter model.Filter

func (c ContainerWrapper) GetName() string  { return c.Name }
func (c ContainerWrapper) GetImage() string { return c.Image }

// Wrapper for corev1.EphemeralContainer
type EphemeralContainerWrapper struct {
	corev1.EphemeralContainer
}

func (c EphemeralContainerWrapper) GetName() string  { return c.Name }
func (c EphemeralContainerWrapper) GetImage() string { return c.Image }

var unnecessaryResources = map[string]struct{}{
	"bindings":                  {},
	"tokenreviews":              {},
	"selfsubjectreviews":        {},
	"subjectaccessreviews":      {},
	"selfsubjectrulesreviews":   {},
	"localsubjectaccessreviews": {},
	"selfsubjectaccessreviews":  {},
	"kcm":                       {},
	"ksh":                       {},
}

// NewClientset takes a path to a kubeconfig file and returns a Kubernetes clientset.
func GetClient() (*K8sClient, error) {

	kubeConfigPath := os.Getenv("KUBECONFIG")
	log.Info().Msgf("Reading kubeconfig path %s", kubeConfigPath)

	if kubeConfigPath == "" {
		kubeConfigPath = os.Getenv("HOME") + "/.kube/config"
	}
	if _, err := os.Stat(kubeConfigPath); os.IsNotExist(err) {
		log.Fatal().Msgf("File does not exist: %s", kubeConfigPath)
	} else if err != nil {
		log.Fatal().Msgf("Error accessing file: %v", err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Printf("Error creating config: %v", err)
		return nil, err
	}

	// Create the clientset from the config.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Err(err).Msg("Error creating clientset")
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Err(err).Msg("Error creating dynamic client")
		return nil, err
	}

	// Suppress API server warnings in Kubernetes client-go
	rest.SetDefaultWarningHandler(rest.NoWarnings{})

	K8sClient := &K8sClient{
		K8sContext:    "default",
		Config:        config,
		Client:        clientset,
		DynamicClient: dynamicClient,
	}
	// Create discovery client
	K8sClient.Discovery, err = discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Err(err).Msg("Failed to create discovery client")
	}
	return K8sClient, nil
}

func (c *K8sClient) GetAllComponents(ctx context.Context) ([]model.Component, []string, error) {
	var namespaces []string
	// Get all API resources
	apiResourceLists, err := c.Discovery.ServerPreferredResources()
	if err != nil {
		log.Err(err).Msg("Failed to list API groups and resources")
	}

	var k8sResourceList []model.Component

	for _, resourceList := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			log.Err(err).Msg("Could not retrieve group version")
		}
		// First, go through all the non namespaced resources, store them, and get the list of namespaces
		for _, resource := range resourceList.APIResources {
			log.Info().Msgf("Processing resource: %s", resource.Name)
			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: resource.Name,
			}
			// Handle pagination while fetching resources
			var continueToken string
			for {
				listOptions := metav1.ListOptions{
					Continue: continueToken, // Use pagination token if present
				}

				k8sResources, k8serr := c.DynamicClient.Resource(gvr).List(ctx, listOptions)
				if k8serr != nil {
					if _, exists := unnecessaryResources[gvr.Resource]; exists {
						log.Debug().Msgf("Failed to list resources for less common resource: %v - error: %v", gvr.Resource, k8serr)
					} else {
						log.Warn().Msgf("Failed to list resources for resource: %v - error: %v", gvr.Resource, k8serr)
					}
					break
				}
				if k8sResources == nil || len(k8sResources.Items) == 0 {
					log.Debug().Msgf("No resources found for GVR: %v", gvr)
					break
				}

				for _, item := range k8sResources.Items {
					namespace := item.GetNamespace()
					if item.GetKind() == "Namespace" {
						namespaces = append(namespaces, item.GetName())
					}
					// For namespaced resources check based on the filter
					if namespace != "" && !K8Filter.ShouldIncludeThisResource(namespace, item.GetKind()) {
						continue
					}

					// For non-namespaced resources
					if !K8Filter.IncludesAllKindsNonNamespaced() {
						if namespace == "" && !utils.Contains(K8Filter.NonNamespacedInclusions.Resources, item.GetKind()) {
							continue
						}
					}
					addToComponentList(item, &k8sResourceList)
				}

				// Handle pagination
				continueToken = k8sResources.GetContinue()
				if continueToken == "" {
					break
				}
			}
		}
	}
	return k8sResourceList, namespaces, nil
}

func (c *K8sClient) GetAllImages(ctx context.Context, namespaceList []string) ([]model.Component, error) {
	imageMap := make(map[string]*model.Component) // A map of the image purl to make sure each one appears only once
	var componentList []*model.Component
	var primaryOwnerRef string
	for _, namespace := range namespaceList {
		pods, err := c.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list pods: %w", err)
		}
		log.Info().Msgf("Listing pods in namespace: %s", namespace)
		for _, pod := range pods.Items {
			ownerReferenceSet := set.Set[string]{}
			if len(pod.OwnerReferences) == 0 {
				log.Info().Msgf("No owner reference found for pod: %s", pod.Name)
			} else {
				//Special cases for Owner References:
				//	1. ReplicaSets can be owned by Deployments or custom CRDs
				//	2. Jobs can be owned by CronJobs or custom CRDs
				primaryOwnerRef = getPrimaryOwnerReference(c, pod.OwnerReferences, &ownerReferenceSet, namespace)
			}

			// Ephemeral containers are added only after the Pod is running but do not affect pod health
			ephemeralContainerStatuses := pod.Status.EphemeralContainerStatuses
			for _, container := range pod.Spec.EphemeralContainers {
				addOrUpdateImageInComponentList(EphemeralContainerWrapper{container}, namespace, &componentList, ephemeralContainerStatuses, "ephemeral", ownerReferenceSet, primaryOwnerRef, imageMap)
			}

			initContainerStatuses := pod.Status.InitContainerStatuses
			for _, container := range pod.Spec.InitContainers {
				addOrUpdateImageInComponentList(ContainerWrapper{container}, namespace, &componentList, initContainerStatuses, "init", ownerReferenceSet, primaryOwnerRef, imageMap)
			}

			containerStatuses := pod.Status.ContainerStatuses
			for _, container := range pod.Spec.Containers {
				addOrUpdateImageInComponentList(ContainerWrapper{container}, namespace, &componentList, containerStatuses, "main", ownerReferenceSet, primaryOwnerRef, imageMap)
			}
		}
	}
	var finalList []model.Component
	for _, compPtr := range componentList {
		finalList = append(finalList, *compPtr) // Dereference pointers before returning
	}
	return finalList, nil
}

func getPrimaryOwnerReference(k *K8sClient, ownerRefs []metav1.OwnerReference, ownerReferenceSet *set.Set[string], namespace string) string {

	//Special cases for Owner References:
	//	1. ReplicaSets can be owned by Deployments or custom CRDs
	//	2. Jobs can be owned by CronJobs or custom CRDs
	ownerReference := getOnePrimaryOwnerReference(ownerRefs)

	//for _, ownerReference := range ownerRefs {
	ownerReferenceKey := fmt.Sprintf("%s/%s", ownerReference.Kind, ownerReference.Name)
	if ownerReferenceSet.Has(ownerReferenceKey) {
		// Already identified the owner reference for this pod
		//continue
		return ownerReferenceKey
	}

	if ownerReference.Kind == "ReplicaSet" { //ReplicaSets can be managed by Deployments
		replicaSetName := ownerReference.Name

		// Retrieve the ReplicaSet
		replicaSet, err := k.Client.AppsV1().ReplicaSets(namespace).Get(context.TODO(), replicaSetName, metav1.GetOptions{})
		if err != nil {
			log.Err(err).Msgf("Error retrieving Replicaset: %s", replicaSetName)
			return ownerReferenceKey
		}

		if len(replicaSet.OwnerReferences) > 0 {
			primaryOwnerRef := getOnePrimaryOwnerReference(replicaSet.OwnerReferences)
			//continue
			ownerReferenceKey = fmt.Sprintf("%s/%s", primaryOwnerRef.Kind, primaryOwnerRef.Name)
		}
	}

	if ownerReference.Kind == "Job" { //Jobs can be managed by CronJobs
		jobName := ownerReference.Name
		job, err := k.Client.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})

		if err != nil {
			log.Err(err).Msgf("Error retrieving Job: %s", jobName)
			return ownerReferenceKey
		}

		if len(job.OwnerReferences) > 0 {
			primaryOwnerRef := getOnePrimaryOwnerReference(job.OwnerReferences)
			//continue
			ownerReferenceKey = fmt.Sprintf("%s/%s", primaryOwnerRef.Kind, primaryOwnerRef.Name)

		}
	}
	ownerReferenceSet.Insert(ownerReferenceKey)
	return ownerReferenceKey
	//}
}

func getOnePrimaryOwnerReference(ownerRefs []metav1.OwnerReference) metav1.OwnerReference {
	// Define priority for Kubernetes-native kinds
	nativeKinds := map[string]int{
		"Deployment":  1,
		"StatefulSet": 2,
		"DaemonSet":   3,
		"CronJob":     4,
	}

	var primaryOwner metav1.OwnerReference
	highestPriority := len(nativeKinds) + 1

	// Check for the highest-priority native kind
	for _, owner := range ownerRefs {
		if priority, exists := nativeKinds[owner.Kind]; exists && priority < highestPriority {
			primaryOwner = owner
			highestPriority = priority
		}
	}

	// Fallback to any custom resource if no native kind is found
	if primaryOwner.Kind == "" && len(ownerRefs) > 0 {
		primaryOwner = ownerRefs[0]
	}

	return primaryOwner
}

func addOrUpdateImageInComponentList(container ContainerLike, namespace string, k8sResourceList *[]*model.Component, containerStatuses []v1.ContainerStatus, source string, ownerRefs set.Set[string], primaryOwnerRef string, imageMap map[string]*model.Component) {
	var properties []model.Property
	var imageId = ""
	var imageSha = ""
	var version = ""
	for _, containerStatus := range containerStatuses {
		if containerStatus.Name == container.GetName() {
			if containerStatus.State.Terminated != nil {
				return // Ignore the container which is already terminated
			}

			imageId = containerStatus.ImageID
			sha256 := "sha256:"
			if strings.Contains(imageId, sha256) {
				imageSha = fmt.Sprintf("%s%s", sha256, strings.Split(imageId, sha256)[1])
			} else {
				log.Error().Msgf("SHA256 digest not found in image: %s - continuing.", imageId)
			}
			break
		}
	}

	component := &model.Component{
		Type:       "container",
		Name:       container.GetImage(), //Pass the full image name and split it into name and version in the function addPropertiesForImageComponent
		PackageURL: "",
		Properties: properties,
		Licenses:   nil,
		Hashes:     nil,
	}
	component.AddProperty(model.ComponentNamespace, namespace)
	component.PackageURL, version = GetImagePkgID(component, imageSha, primaryOwnerRef)

	if c, exists := imageMap[component.PackageURL]; exists {
		updateImageInComponentList(c, source, ownerRefs, primaryOwnerRef, version)
		log.Debug().Msgf("Updated existing image for resource: %s, kind: image, namespace: %s", container.GetImage(), namespace)
	} else {
		c = addImageToComponentList(component, namespace, k8sResourceList, source, ownerRefs, primaryOwnerRef, imageMap)
		log.Debug().Msgf("Added new image for resource: %s, kind: image, namespace: %s", container.GetImage(), namespace)
		//add to imagemap
		imageMap[component.PackageURL] = c
	}
}

func updateImageInComponentList(component *model.Component, source string, ownerRefs set.Set[string], primaryOwner string, version string) {
	prop, exists := component.GetPropertyObject(model.ComponentOwnerRef)
	//Note: Handle multiple containers owner references properly, currently assuming single primary owner reference
	prop, exists = component.GetPropertyObject(model.ComponentSourceRef)
	if exists {
		prop.Values = []string{source}
	} else {
		component.AddProperty("clx:k8s:source", source)
	}
	prop, exists = component.GetPropertyObject(model.ComponentVersion)
	if exists {
		prop.Values = []string{version}
	} else {
		component.AddProperty(model.ComponentVersion, version)
	}
}

func addImageToComponentList(component *model.Component, namespace string, k8sResourceList *[]*model.Component, source string, ownerRefs set.Set[string], primaryOwner string, imageMap map[string]*model.Component) *model.Component {
	// Add a new component
	imageMap[component.PackageURL] = component
	component.AddProperty(model.ComponentKind, "Image")
	component.AddProperty(model.ComponentNamespace, namespace)
	component.AddProperty("clx:k8s:source", source)
	component.AddProperty("clx:k8s:ownerRef", primaryOwner)
	component.AddProperty(model.ComponentVersion, component.Version)

	//TODO: capture multipe ownerRefs in pedigree or externalReferences property in component

	*k8sResourceList = append(*k8sResourceList, component)
	log.Debug().Msgf("Created new image for resource: %s, kind: image, namespace: %s", component.Name, namespace)
	return component
}

func addToComponentList(item unstructured.Unstructured, k8sResourceList *[]model.Component) {
	var properties []model.Property
	component := model.Component{
		Type:       "application",
		Name:       item.GetName(),
		Version:    item.GetAPIVersion(),
		PackageURL: "",
		Properties: properties,
		Licenses:   nil,
		Hashes:     nil,
	}

	component.AddProperty(model.ComponentKind, item.GetKind())
	component.AddProperty(model.ComponentNamespace, item.GetNamespace())
	addVersionForComponent(item, &component, "clx:k8s:componentVersion")
	component.PackageURL = GetAppPkgId(item.GetKind(), item.GetName(), item.GetNamespace(), item.GetAPIVersion())
	*k8sResourceList = append(*k8sResourceList, component)
	log.Debug().Msgf("Created new component for resource: %s, kind: %s, namespace: %s", item.GetName(), item.GetKind(), item.GetNamespace())
}

func GetAppPkgId(kind string, name string, namespace string, apiVersion string) string {
	baseUrl := fmt.Sprintf("%s:%s/%s/%s", model.PkgPrefix, model.K8sPrefix, kind, name)
	urlValues := url.Values{
		"apiVersion": []string{apiVersion},
	}
	// Some resources don't have a namespace, only add to the purl if namespace exists
	if namespace != "" {
		urlValues.Add("namespace", namespace)
	}

	//Format:  pkg:kubernetes/{kind}/{name}?apiVersion={apiVersion}&namespace={namespace}
	return fmt.Sprintf("%s?%s", baseUrl, urlValues.Encode())
}

func addVersionForComponent(item unstructured.Unstructured, component *model.Component, key string) {
	componentKind := item.GetKind()
	// Get the version based on component kind since there is no standard way of setting component's version in custom resources
	switch componentKind {
	case "HelmChart":
		componentSpec, ok := item.Object["spec"].(map[string]interface{})
		componentVersion, err := componentSpec["version"].(string)
		if !ok || !err {
			log.Info().Msgf("Fetching version from label helm.sh/chart for component: %s", component.Name)
			addLabelIfExists(item, "helm.sh/chart", component, "clx:k8s:componentVersion")
		} else {
			component.AddProperty(key, componentVersion)
		}
	default:
		addLabelIfExists(item, "helm.sh/chart", component, "clx:k8s:componentVersion")
	}
}

func addLabelIfExists(item unstructured.Unstructured, label string, component *model.Component, propertyKey string) {
	// Get labels map safely
	labels, ok := item.Object["metadata"].(map[string]interface{})["labels"].(map[string]interface{})
	if !ok {
		return
	}

	// Get the label safely
	labelValue, exists := labels[label]
	if !exists {
		log.Debug().Msgf("Item %s does not have a label %s", item.GetName(), label)
		return
	}

	// Ensure it's a string before returning
	labelValueStr, valid := labelValue.(string)
	if !valid {
		log.Error().Msgf("Item %s has a label %s with value %v which is not a string.", item.GetName(), label, labelValue)
		return
	}

	component.AddProperty(propertyKey, labelValueStr)
}

func GetImagePkgID(imageComponent *model.Component, imageSha string, primaryOwnerRef string) (string, string) {
	ref, err := name.ParseReference(imageComponent.Name)
	if err != nil {
		log.Err(err).Msgf("No reference found for Image: %s", imageComponent.Name)
	} else {
		imageComponent.Version = ref.Identifier()
		imageComponent.Name = strings.Split(ref.Name(), ":")[0]
	}

	baseUrl := ref.Context().RepositoryStr()

	baseName := fmt.Sprintf("%s:%s/%s", model.PkgPrefix, model.OciPrefix, baseUrl)

	urlValues := url.Values{
		"repository_url": []string{imageComponent.Name},
	}

	urlValues.Add("ownerRef", primaryOwnerRef)
	urlValues.Add("namespace", imageComponent.GetNamespace())
	if imageSha != "" {
		baseName = fmt.Sprintf("%s@%s", baseName, imageSha)
	}

	//Format:  pkg:oci/{imageName}/{@ImageSha}?namespace={namespace}&ownerRef={primaryOwnerRef}&repository_url={repourl}&
	return fmt.Sprintf("%s?%s", baseName, urlValues.Encode()), imageComponent.Version
}
