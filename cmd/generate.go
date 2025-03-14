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
	format  string
	outPath string
	filters string
	sort    bool
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Kubernetes BOM for the provided K8s cluster",
	RunE:  runGenerate,
}

const defaultFilterFileName = "./filters.json"

func init() {
	GenerateCmd.Flags().StringVarP(&format, "format", "f", "cyclonedx-json", "Format of the generated BOM.")
	GenerateCmd.Flags().StringVarP(&outPath, "out-path", "o", "./output.json", "Path and filename of generated cluster codex file.")
	GenerateCmd.Flags().StringVarP(&filters, "filters", "i", "", "Path to a json file containing inclusion filters. (default file name: filters.json)")
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

	componentList, err := k8client.GetAllComponents(ctx)
	if err != nil {
		config.ClxLogger.Error("Error getting resources.", "error", err)
		return nil // Handle error case by returning nil BOM or some default value
	}
	bom.Components = componentList

	namespaceList := getNamespaceList()
	if len(namespaceList) <= 0 {
		namespaceComponents := bom.FindApplicationsByKind("Namespace", "")
		for _, component := range namespaceComponents {
			namespaceList = append(namespaceList, component.Name)
		}
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

func getInclusionFilter() error {
	k8.K8Filter = &model.Inclusions{Inclusions: []model.Inclusion{}}
	if filters != "" {
		// Check if file exists
		if _, err := os.Stat(filters); os.IsNotExist(err) {
			if filters == defaultFilterFileName {
				config.ClxLogger.Debug("Default filter file did not exist.")
				return nil
			}
			return fmt.Errorf("file does not exist: %s", filters)
		}

		// Read file contents
		data, err := ioutil.ReadFile(filters)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}

		// Parse JSON
		var inc model.Inclusions
		err = json.Unmarshal(data, &inc)
		if err != nil {
			return fmt.Errorf("failed to parse JSON: %v", err)
		}

		for idx, inclusion := range inc.Inclusions {
			// Convert Resources to lowercase
			for j, resource := range inclusion.Resources {
				inc.Inclusions[idx].Resources[j] = strings.ToLower(resource)
			}
		}

		k8.K8Filter = &inc
	}
	return nil
}

func getNamespaceList() []string {

	var namespaces []string
	for _, inclusion := range k8.K8Filter.Inclusions {
		namespaces = append(namespaces, inclusion.Namespace)
	}
	return namespaces
}
