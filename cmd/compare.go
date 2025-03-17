package cmd

import (
	"encoding/json"
	"github.com/golang-collections/collections/set"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"regexp"
)

var (
	expectedKBOMPath string
	actualKBOMPath   string
)

const APPLICATION = "application"
const CONTAINER = "container"
const BOM_PROPERTY_NAMESPACE = "clx:k8s:componentNamespace"
const ComponentType = "component-type"

const ECR_REGEX = `(\d+)\.dkr\.ecr\.[a-z0-9-]+\.amazonaws\.com`

var ecrRegex = regexp.MustCompile(ECR_REGEX)

var CompareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare two Kubernetes BOM files against one another",
	Long: `Compare two Kubernetes BOM files against one another, and return any diff between the two files.

	Example usage: clx compare --expected=golden_kbom.json --actual=new_kbom.json`,
	RunE: compareKBOM,
}

func init() {
	CompareCmd.Flags().StringVarP(&expectedKBOMPath, "expected", "e", "", "Filepath to the golden Kubernetes BOM (ie the source of truth)")
	CompareCmd.MarkFlagRequired("expected")
	CompareCmd.Flags().StringVarP(&actualKBOMPath, "actual", "a", "", "Filepath to the Kubernetes BOM to be compared against")
	CompareCmd.MarkFlagRequired("actual")
}

func compareKBOM(cmd *cobra.Command, _ []string) error {

	expected, err := os.ReadFile(expectedKBOMPath)
	if err != nil {
		return err
	}
	actual, err := os.ReadFile(actualKBOMPath)
	if err != nil {
		return err
	}

	var expectedStruct *unstructured.Unstructured // Golden KBOM
	var actualStruct *unstructured.Unstructured   // Generated KBOM on given cluster

	json.Unmarshal(expected, &expectedStruct)
	json.Unmarshal(actual, &actualStruct)

	// Convert JSON to MAP for the container and application type
	expectedContainerMap := extractBOMToMap(expectedStruct, CONTAINER)
	expectedApplicationMap := extractBOMToMap(expectedStruct, APPLICATION)

	actualContainerMap := extractBOMToMap(actualStruct, CONTAINER)
	actualApplicationMap := extractBOMToMap(actualStruct, APPLICATION)

	//for key, value := range expectedContainerMap {
	//	fmt.Printf("- %s: %d\n", key, value)
	//}

	return nil
}

func extractBOMToMap(data *unstructured.Unstructured, dataType string) map[string]map[string]string {
	dataMap := make(map[string]map[string]string)

	for _, value := range data.Object["components"].([]interface{}) {
		component := value.(map[string]interface{})
		kbomType := component["type"].(string)
		purl := component["purl"].(string)
		properties := component["properties"].([]interface{})

		innerContainerMap := make(map[string]string)

		if kbomType == dataType {
			for _, props := range properties {
				itemKey := props.(map[string]interface{})["name"].(string)
				itemValue := props.(map[string]interface{})["value"].(string)
				innerContainerMap[itemKey] = itemValue
			}
			dataMap[purl] = innerContainerMap
		}
	}

	return dataMap
}
