package cmd

import (
	"cluster-codex/internal/k8"
	"cluster-codex/internal/model"
	"context"
	prettyjson "encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/version"
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
	bom := GenerateBOM(k8sClient)
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

// func GenerateBOM(k8sClient *k8.K8sClient) *model.BOM {

func GenerateBOM(k8client k8.K8sClientInterface) *model.BOM {

	bom := model.NewBOM()
	ctx := context.Background()

	componentList, err := k8client.GetAllComponents(ctx)
	if err != nil {
		log.Printf("Error getting resources: %v", err)
		return nil // Handle error case by returning nil BOM or some default value
	}

	bom.Components = componentList

	return bom
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
