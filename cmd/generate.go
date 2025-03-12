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
	GenerateCmd.Flags().StringVarP(&filters, "filters", "i", defaultFilterFileName, "Path to a json file containing inclusion filters.")
	GenerateCmd.Flags().BoolVarP(&sort, "sort", "s", false, "Sort the generated BOM JSON in Application, Kind, Name, Namespace order")
}

func runGenerate(cmd *cobra.Command, _ []string) error {
	config.ClxLogger.Info("Starting generate command\n")
	start := time.Now()

	// Read filter file, if any.
	var namespaces []string
	var err error
	namespaces, err = GetNamespacesFromJSON(filters)
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

	bom := GenerateBOM(k8sClient, namespaces)

	// Sort the BOM so it is consistent
	if sort {
		bom.Sort()
	}

	// Output the BOM as JSON
	jsonData, err := prettyjson.MarshalIndent(bom, "", "  ")
	if err != nil {
		config.ClxLogger.Error("Got error converting output BOM to json %v", "error", err)
		os.Exit(1)
	}

	// Create a file and write the json
	err = ValidatePath(outPath)
	if err != nil {
		config.ClxLogger.Error("Error validating path: %v", "error", err)
	}
	file, err := os.Create(outPath)
	if err != nil {
		config.ClxLogger.Error("Error creating file %s: %v\n", outPath, err)
		os.Exit(1)
	}
	defer file.Close()

	// Write JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Printf("Error writing to file %s: %v\n", outPath, err)
		return err
	}
	elapsed := time.Since(start)
	rounded := elapsed.Round(time.Second)
	seconds := int64(rounded / time.Second)
	fmt.Printf("Generate command output written to file %s in %d seconds\n", outPath, seconds)
	return err
}

func GenerateBOM(k8client k8.K8sClientInterface, namespaces []string) *model.BOM {

	bom := model.NewBOM()

	var ctx context.Context
	if len(namespaces) > 0 {
		ctx = context.WithValue(context.Background(), k8.FilterKey, k8.K8sFilter{
			Namespaces: namespaces,
		})
	} else {
		ctx = context.Background()
	}

	componentList, err := k8client.GetAllComponents(ctx)
	if err != nil {
		config.ClxLogger.Error("Error getting resources.", "error", err)
		return nil // Handle error case by returning nil BOM or some default value
	}
	bom.Components = componentList

	var namespaceList []string
	// If we are filtering by namespace, only look for images in those namespaces
	if len(namespaces) > 0 {
		namespaceList = namespaces
	} else {
		// Otherwise look for images in all namespaces
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
