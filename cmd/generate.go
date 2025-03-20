package cmd

import (
	"cluster-codex/internal/config"
	"cluster-codex/internal/k8"
	"cluster-codex/internal/model"
	"context"
	prettyjson "encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/version"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	format     string
	outPath    string
	filterPath string
	sort       bool
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Kubernetes BOM for the provided K8s cluster",
	RunE:  runGenerate,
}

func init() {
	GenerateCmd.Flags().StringVarP(&format, "format", "f", "cyclonedx-json", "Format of the generated BOM.")
	GenerateCmd.Flags().StringVarP(&outPath, "out-path", "o", "./output.json", "Path and filename of generated cluster codex file.")
	GenerateCmd.Flags().StringVarP(&filterPath, "filter-path", "i", "", "Path to a json file containing inclusion filterPath.")
	GenerateCmd.Flags().BoolVarP(&sort, "sort", "s", false, "Sort the generated BOM JSON in Application, Kind, Name, Namespace order")
}

func runGenerate(cmd *cobra.Command, _ []string) error {
	config.ClxLogger.Info("Starting generate command\n")
	start := time.Now()

	// Read filter file, if any.
	var err error
	err = getInclusionFilter()
	if err != nil {
		//config.ClxLogger.Error("Error loading filter file", "error", err)
		//os.Exit(1)
		config.ClxLogger.Error("Error loading filter file", "error", err)
		log.Fatalf("Error loading filter file: %v", err)
	}

	k8sClient, err := k8.GetClient()

	if err != nil {
		config.ClxLogger.Error("Error creating Kubernetes client: %v", "error", err)
		os.Exit(1)
	}
	var serverVersion *version.Info
	serverVersion, err = k8sClient.Client.Discovery().ServerVersion()
	if err != nil {
		config.ClxLogger.Error("Failed to get server version: %v", "error", err)
		os.Exit(1)
	}

	config.ClxLogger.Info("Git:", "Version", serverVersion.String())

	bom := GenerateBOM(k8sClient)

	// Sort the BOM so it is consistent
	if sort {
		bom.Sort()
	}

	err = writeJson(bom)
	if err != nil {
		return err
	}

	elapsed := time.Since(start)
	rounded := elapsed.Round(time.Second)
	seconds := int64(rounded / time.Second)
	fmt.Printf("Generate command output written to file %s in %d seconds\n", outPath, seconds)
	return err
}

func writeJson(bom *model.BOM) error {
	err := ValidatePath(outPath)
	if err != nil {
		log.Fatalf("Error validating path: %v", err)
	}
	file, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", outPath, err)
		return err
	}
	defer file.Close()

	encoder := prettyjson.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ") // Equivalent to MarshalIndent

	if err := encoder.Encode(bom); err != nil {
		log.Fatalf("Got error converting to json %v", err)
	}
	return err
}

func GenerateBOM(k8client k8.K8sClientInterface) *model.BOM {

	bom := model.NewBOM()
	ctx := context.Background()

	componentList, namespaces, err := k8client.GetAllComponents(ctx)
	if err != nil {
		config.ClxLogger.Error("Error getting resources.", "error", err)
		return nil // Handle error case by returning nil BOM or some default value
	}
	bom.Components = componentList

	namespaceList := k8.K8Filter.GetNamespaceList()
	if len(namespaceList) <= 0 {
		namespaceList = namespaces // Get the list of namespaces if no filter is defined for namespaces
	}

	componentList, err = k8client.GetAllImages(ctx, namespaceList)
	if err != nil {
		config.ClxLogger.Error("Error getting images.", "error", err)
		return nil // Handle error case by returning nil BOM or some default value
	}
	bom.Components = append(bom.Components, componentList...)
	return bom
}

func ValidatePath(filePath string) error {
	if filePath == "" {
		return errors.New("path cannot be empty")
	}

	invalidChars := []string{"|", "<", ">", "?", "*", ":", "\\"}
	for _, char := range invalidChars {
		if strings.Contains(filePath, char) {
			return errors.New("path contains invalid character: " + char)
		}
	}

	if filepath.Ext(filePath) == "" {
		return errors.New("path must have a file extension")
	}

	return nil
}

// TODO: Add non-namespaced as well
func getInclusionFilter() error {
	//k8.K8Filter = &model.Inclusions{Inclusions: []model.Inclusion{{Namespaces: []string{""}}}}
	//var filter = model.Filter{NamespacedInclusions: []model.NamespacedInclusion{{Namespaces: []string{"*"}}}}

	if filterPath != "" {
		// Check if file exists
		if _, err := os.Stat(filterPath); os.IsNotExist(err) {
			config.ClxLogger.Error("file does not exist", "filterPath file path", filterPath)
			log.Fatalf("file does not exist: %s", err)
		}

		// Read file contents
		data, err := ioutil.ReadFile(filterPath)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}

		// Parse JSON
		var filter model.Filter
		err = json.Unmarshal(data, &filter)
		if err != nil {
			return fmt.Errorf("failed to parse JSON: %v", err)
		}

		//for idx, inclusion := range filter.NamespacedInclusions {
		//	//set default namespace to "*" if it's not provided in the filter'
		//	if (inclusion.Namespaces == nil) && len(inclusion.Namespaces) == 0 {
		//		filter.NamespacedInclusions[idx].Namespaces = []string{"*"}
		//	}
		//	// Convert Resources to lowercase
		//	for j, resource := range inclusion.Resources {
		//		filter.NamespacedInclusions[idx].Resources[j] = strings.ToLower(resource)
		//	}
		//}

		// The below logic is to detect
		// * if the namespace is "*" and the corresponding resource array is empty considered for all resources for all namespaces.
		// * if the namespace is "*" and the corresponding resource array is NOT empty the namespace or that inclusion set to "*" which means query the specific resources in all namespaces
		for idx, inclusion := range filter.NamespacedInclusions {
			//set default namespace to "*" if it's not provided in the filter'
			if (inclusion.Namespaces == nil) && len(inclusion.Namespaces) == 0 {
				filter.NamespacedInclusions[idx].Namespaces = []string{"*"}
			}
			for j, _ := range inclusion.Namespaces {
				if filter.NamespacedInclusions[idx].Namespaces[j] == "*" {
					filter.NamespacedInclusions[idx].Namespaces = []string{"*"}
					if len(filter.NamespacedInclusions[idx].Resources) == 0 {
						// If resource array is empty, consider all resources for all namespaces
						filter.NamespacedInclusions[idx].Resources = []string{"*"}
					}
					// Convert Resources to lowercase
					for k, resource := range inclusion.Resources {
						filter.NamespacedInclusions[idx].Resources[k] = strings.ToLower(resource)
					}
					break
				}
			}
		}
		k8.K8Filter = filter
	}
	return nil
}
