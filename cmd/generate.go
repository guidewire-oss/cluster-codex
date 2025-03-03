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
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Kubernetes BOM for the provided K8s cluster",
	RunE:  runGenerate,
}

func init() {
	GenerateCmd.Flags().StringVarP(&format, "format", "f", "cyclonedx-json", "Format of the generated BOM.")
	GenerateCmd.Flags().StringVarP(&outPath, "out-path", "o", "./output.json", "Path and filename of generated cluster codex file.")

}

func runGenerate(cmd *cobra.Command, _ []string) error {
	fmt.Printf("Starting generate command\n")
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

	config.ClxLogger.Info("Git:", "Version", serverVersion.String())

	bom := GenerateBOM(k8sClient)
	jsonData, err := prettyjson.MarshalIndent(bom, "", "  ")
	if err != nil {
		log.Fatalf("Got error converting to json %v", err)
	}

	// Create a file and write the json
	err = ValidatePath(outPath)
	if err != nil {
		log.Fatalf("Error validating path: %v", err)
	}
	file, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", outPath, err)
		return err
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

func GenerateBOM(k8client k8.K8sClientInterface) *model.BOM {

	bom := model.NewBOM()
	ctx := context.Background()

	componentList, err := k8client.GetAllComponents(ctx)
	if err != nil {
		config.ClxLogger.Error("Error getting resources.", "error", err)
		return nil // Handle error case by returning nil BOM or some default value
	}

	bom.Components = componentList
	var namespaceList []string
	namespaceComponents := bom.FindApplicationsByKind("Namespace", "")

	for _, component := range namespaceComponents {
		namespaceList = append(namespaceList, component.Name)
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
