package cmd

import (
	"cluster-codex/internal/k8"
	"cluster-codex/internal/model"
	"context"
	prettyjson "encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	format  string
	outPath string
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate KBOM for the provided K8s cluster",
	RunE:  runGenerate,
}

func init() {
	GenerateCmd.Flags().StringVarP(&format, "format", "f", "cyclonedx-json", "Format of the generated BOM.")
	GenerateCmd.Flags().StringVarP(&outPath, "out-path", "o", ".", "Path and filename to write cluster codex file to.")

	//Nice to have: bind flags to Viper: https://pkg.go.dev/github.com/spf13/pflag#section-readme
	// viper.Set("format",...)
}

func runGenerate(cmd *cobra.Command, _ []string) error {
	start := time.Now()
	k8sClient, err := k8.GetClient()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}
	var serverVersion *version.Info
	serverVersion, err = k8sClient.Client.Discovery().ServerVersion()
	if err != nil {
		log.Fatalf("Failed to get server version: %v", err)
	}

	fmt.Printf("Git version:%s\n", serverVersion.String())

	//testing section
	//config, err := rest.InClusterConfig()
	//if err != nil {
	//	log.Fatalf("Failed to load in-cluster config: %v", err)
	//}
	bom := generateBOM(k8sClient)
	jsonData, err := prettyjson.MarshalIndent(bom, "", "  ")
	if err != nil {
		log.Fatalf("Got error converting to json %v", err)
	}
	log.Printf("Final KBOM: %v", string(jsonData))

	//Create a fiel and write the json
	file, err := os.Create("output.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	// Write JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	elapsed := time.Since(start)
	rounded := elapsed.Round(time.Second)
	seconds := int64(rounded / time.Second)
	fmt.Printf("generate command took %d seconds\n", seconds)
	return err
}

func generateBOM(k8sClient *k8.K8sClient) *model.BOM {
	bom := model.NewBOM()

	// Create discovery client
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(k8sClient.Config)
	if err != nil {
		log.Fatalf("Failed to create discovery client: %v", err)
	}

	componentList, err := AllApiResources(err, discoveryClient, k8sClient)
	if err != nil {
		log.Printf("Error getting resources")
	}

	bom.Components = componentList

	return bom
}

func AllApiResources(err error, discoveryClient *discovery.DiscoveryClient, k8sClient *k8.K8sClient) ([]model.Component, error) {

	ctx := context.Background()
	// Get all API resources
	//apiGroups, apiResourceLists, err := discoveryClient.ServerGroupsAndResources()

	apiResourceLists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		log.Fatalf("Failed to list API groups and resources: %v", err)
	}

	//var apiResourceComponents []model.Component
	//var k8sResourceList []*unstructured.Unstructured
	var k8sResourceList []model.Component

	for _, resourceList := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			log.Fatalf("Could not retrieve group version %v", err)
		}
		namespaces := []string{}
		// First, go through all the non namespaced resources, store them, and get the list of namespaces
		for _, resource := range resourceList.APIResources {
			if resource.Namespaced {
				continue
			}
			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: resource.Name,
			}
			k8sResources, k8serr := k8sClient.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
			if k8serr != nil {
				log.Printf("Failed to list resources for %v %v", gvr, k8serr)
				continue
			}
			for _, item := range k8sResources.Items {
				if resource.Name == "namespace" {
					namespaces = append(namespaces, item.GetName())
				}
				k8sResourceList = append(k8sResourceList, createComponentList(item, "", k8sResourceList)...)
			}
		}

		// Next, go through all the namespaced resources, and for each one, get all the resources in each namespace.
		for _, resource := range resourceList.APIResources {
			if !resource.Namespaced {
				continue
			}
			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: resource.Name,
			}
			for _, namespace := range namespaces {
				k8sResources, k8serr := k8sClient.DynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
				if k8serr != nil {
					log.Printf("Failed to list resources in namespace %v %s %v", gvr, namespace, k8serr)
					continue
				}

				if k8sResources == nil || len(k8sResources.Items) == 0 {
					log.Printf("No resources found for GVR %v in namespace %s", gvr, namespace)
					continue
				}

				// Append the items to the resourceList slice
				// We need to create a copy of the item before appending otherwise if we append &item
				// we are appending pointers to the same underlying objects from resources.Items
				for _, item := range k8sResources.Items {
					k8sResourceList = append(k8sResourceList, createComponentList(item, namespace, k8sResourceList)...)
				}
			}
		}

		//
		//	//TODO:update this to get all namespaces
		//	// namespaces := []string{"atmos-system"}
		//	for _, namespace := range namespaces {
		//		//k8sResources, k8serr := k8sClient.DynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
		//		k8sResources, k8serr := k8sClient.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
		//		if k8serr != nil {
		//			log.Printf("Failed to list resources in namespace %v %s %v", gvr, namespace, k8serr)
		//			continue
		//		}
		//
		//		if k8sResources == nil || len(k8sResources.Items) == 0 {
		//			log.Printf("No resources found for GVR %v in namespace %s", gvr, namespace)
		//			continue
		//		}
		//
		//		// Append the items to the resourceList slice
		//		// We need to create a copy of the item before appending otherwise if we append &item
		//		// we are appending pointers to the same underlying objects from resources.Items
		//		for _, item := range k8sResources.Items {
		//			k8sResourceList = createComponentList(item, namespace, k8sResourceList)
		//		}
		//	}
		//}

		//for _, resource := range resourceList.APIResources {
		//	component := model.Component{
		//		Type:    resource.Kind,
		//		Name:    resource.Name,
		//		Version: resource.Version,
		//	}
		//	fmt.Printf("\nCurrent resource: %v\n", resource)
		//
		//	component.PackageURL = PkgID(component) //TODO: fill out remaining Component fields
		//
		//	apiResourceComponents = append(apiResourceComponents, component)
		//}
	}

	return k8sResourceList, nil
}

func createComponentList(item unstructured.Unstructured, namespace string, k8sResourceList []model.Component) []model.Component {
	var properties = []model.Property{}
	component := model.Component{
		Type:       "application",
		Name:       item.GetName(),
		Version:    item.GetAPIVersion(),
		PackageURL: "",
		Properties: properties,
		Licenses:   nil,
		Hashes:     nil,
	}

	namespaceTemp := "No Namespace"
	if namespace != "" {
		namespaceTemp = namespace
	}

	component.AddProperty("clx:k8s:componentKind", item.GetKind())
	component.AddProperty("clx:k8s:componentApiVersion", item.GetAPIVersion())
	component.AddProperty("clx:k8s:namespace", namespaceTemp)
	addLabelIfExists(item, "helm.sh/chart", &component, "clx:k8s:componentVersion")

	k8sResourceList = append(k8sResourceList, component)
	log.Printf("Current resource: %s %s %s", item.GetName(), item.GetKind(), namespaceTemp)
	return k8sResourceList
}

func addLabelIfExists(item unstructured.Unstructured, label string, component *model.Component, propertyKey string) {
	// Get labels map safely
	labels, ok := item.Object["metadata"].(map[string]interface{})["labels"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: Labels not found or invalid format")
		return
	}

	// Get the label safely
	version, exists := labels[label]
	if !exists {
		fmt.Printf("\nLabel %s not found\n", label)
		return
	}

	// Ensure it's a string before returning
	versionStr, valid := version.(string)
	if !valid {
		fmt.Printf("\nError: label %s is not a string\n", label)
		return
	}

	component.AddProperty(propertyKey, versionStr)
}

func PkgID(component model.Component) string {
	parts := strings.Split(component.Name, "/")
	baseName := fmt.Sprintf("%s:%s/%s", model.PkgPrefix, model.OciPrefix, parts[len(parts)-1])

	urlValues := url.Values{
		"repository_url": []string{component.Name},
	}

	if component.Version != "" {
		urlValues.Add("tag", component.Version)
	}

	//TODO: finish logic for generating PURL
	//if component.OwnerReference != "" {
	//	urlValues.Add("owner", component.OwnerReference)
	//}
	//
	//if component.Digest != "" {
	//	baseName = fmt.Sprintf("%s@%s", baseName, url.QueryEscape(component.Digest))
	//}

	return fmt.Sprintf("%s?%s", baseName, urlValues.Encode())
}
