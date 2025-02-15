package k8

import (
	"cluster-codex/internal/config"
	"cluster-codex/internal/model"
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . K8sClientInterface
type K8sClientInterface interface {
	GetAllComponents(ctx context.Context) ([]model.Component, error)
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

// NewClientset takes a path to a kubeconfig file and returns a Kubernetes clientset.
func GetClient() (*K8sClient, error) {

	kubeConfigPath := os.Getenv("KUBECONFIG")
	fmt.Printf("Reading kubeconfig from %s\n", kubeConfigPath)

	if kubeConfigPath == "" {
		kubeConfigPath = os.Getenv("HOME") + "/.kube/config"
	}
	if _, err := os.Stat(kubeConfigPath); os.IsNotExist(err) {
		log.Fatalf("File does not exist: %s", kubeConfigPath)
	} else if err != nil {
		log.Fatalf("Error accessing file: %v", err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Printf("Error creating config: %v", err)
		return nil, err
	}

	// Create the clientset from the config.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Error creating clientset: %v", err)
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Printf("Error creating dynamic client: %v", err)
		return nil, err
	}

	K8sClient := &K8sClient{
		K8sContext:    "default",
		Config:        config,
		Client:        clientset,
		DynamicClient: dynamicClient,
	}
	// Create discovery client
	K8sClient.Discovery, err = discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create discovery client: %v", err)
	}
	return K8sClient, nil
}

func (c *K8sClient) GetAllComponents(ctx context.Context) ([]model.Component, error) {
	// Get all API resources
	apiResourceLists, err := c.Discovery.ServerPreferredResources()
	if err != nil {
		log.Fatalf("Failed to list API groups and resources: %v", err)
	}

	var k8sResourceList []model.Component

	for _, resourceList := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			log.Fatalf("Could not retrieve group version %v", err)
		}
		// First, go through all the non namespaced resources, store them, and get the list of namespaces
		for _, resource := range resourceList.APIResources {
			config.ClxLogger.Info("Processing resource", "resource", resource.Name)
			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: resource.Name,
			}
			k8sResources, k8serr := c.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
			if k8serr != nil {
				config.ClxLogger.Error("Failed to list resources for", "resource", gvr.Resource, "error", k8serr)
				continue
			}
			if k8sResources == nil || len(k8sResources.Items) == 0 {
				config.ClxLogger.Info("No resources found for GVR", "gvr", gvr)
				continue
			}
			for _, item := range k8sResources.Items {
				addToComponentList(item, &k8sResourceList)
			}
		}
	}

	return k8sResourceList, nil
}

func (c *K8sClient) GetAllImages(ctx context.Context, namespaceList []string) ([]model.Component, error) {
	var componentList []model.Component
	for _, namespace := range namespaceList {
		pods, err := c.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list pods: %w", err)
		}
		config.ClxLogger.Info("Listing pods", "namespace", namespace, "pods", pods.Items)
		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				addImageToComponentList(container, namespace, &componentList)
			}
		}
	}
	return componentList, nil
}

func addImageToComponentList(container v1.Container, namespace string, k8sResourceList *[]model.Component) {
	var properties []model.Property
	component := model.Component{
		Type:       "container",
		Name:       container.Image,
		Version:    container.Image,
		PackageURL: "",
		Properties: properties,
		Licenses:   nil,
		Hashes:     nil,
	}

	component.AddProperty("clx:k8s:componentKind", "Image")
	component.AddProperty("clx:k8s:namespace", namespace)
	//addVersionForComponent(item, &component, "clx:k8s:componentVersion")

	*k8sResourceList = append(*k8sResourceList, component)
	//config.ClxLogger.Info("Created new image for resource:", "name", item.GetName(), "kind", item.GetKind(), "namespace", item.GetNamespace())
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

	component.AddProperty("clx:k8s:componentKind", item.GetKind())
	component.AddProperty("clx:k8s:namespace", item.GetNamespace())
	addVersionForComponent(item, &component, "clx:k8s:componentVersion")

	*k8sResourceList = append(*k8sResourceList, component)
	config.ClxLogger.Info("Created new component for resource:", "name", item.GetName(), "kind", item.GetKind(), "namespace", item.GetNamespace())
}

func addVersionForComponent(item unstructured.Unstructured, component *model.Component, key string) {
	componentKind := item.GetKind()
	// Get the version based on component kind since there is no standard way of setting component's version in custom resources
	switch componentKind {
	case "HelmChart":
		componentSpec, ok := item.Object["spec"].(map[string]interface{})
		componentVersion, err := componentSpec["version"].(string)
		if !ok || !err {
			log.Println("Fetching version from label helm.sh/chart ")
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
		// fmt.Printf("Info: Label %s not found for %s\n", label, item.GetName())
		return
	}

	// Ensure it's a string before returning
	labelValueStr, valid := labelValue.(string)
	if !valid {
		config.ClxLogger.Error("Error: label is not a string for item", "label", label, "item", item.GetName())
		return
	}

	component.AddProperty(propertyKey, labelValueStr)
}
