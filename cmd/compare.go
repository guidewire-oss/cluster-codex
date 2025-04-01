package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"regexp"
)

var (
	expectedBOMPath string
	actualBOMPath   string
)

var (
	red   = "\033[31m"
	green = "\033[32m"
	reset = "\033[0m"
)

type BOMMap struct {
	ContainerMap   map[string]map[string]string
	ApplicationMap map[string]map[string]string
}

const APPLICATION = "application"
const CONTAINER = "container"

const BOM_PROPERTY_NAME = "clx:k8s:component:name"
const BOM_PROPERTY_CONTAINER_NAMESPACE = "clx:k8s:componentNamespace"
const BOM_PROPERTY_VERSION = "clx:k8s:component:version"
const BOM_PROPERTY_OWNERREF = "clx:k8s:component:ownerRef"
const COMPONENT_TYPE = "component-type"

const ECR_REGEX = `(\d+)\.dkr\.ecr\.[a-z0-9-]+\.amazonaws\.com`

var ecrRegex = regexp.MustCompile(ECR_REGEX)

type ComponentData struct {
	Name       string
	Type       string
	Properties []ComparisonProperty
}

type ComparisonProperty struct {
	PropertyName string
	Actual       string
	Expected     string
}

var CompareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare two Kubernetes BOM files against one another",
	Long: `Compare two Kubernetes BOM files against one another, and return any diff between the two files.

	Example usage: clx compare --expected=expected_bom.json --actual=new_bom.json`,
	RunE: compare,
}

func init() {
	CompareCmd.Flags().StringVarP(&expectedBOMPath, "expected", "e", "", "Filepath to the golden Kubernetes BOM (ie the source of truth)")
	CompareCmd.MarkFlagRequired("expected")
	CompareCmd.Flags().StringVarP(&actualBOMPath, "actual", "a", "", "Filepath to the Kubernetes BOM to be compared against")
	CompareCmd.MarkFlagRequired("actual")
}

func compare(cmd *cobra.Command, _ []string) error {

	expected, err := os.ReadFile(expectedBOMPath)
	if err != nil {
		return err
	}
	actual, err := os.ReadFile(actualBOMPath)
	if err != nil {
		return err
	}

	var expectedStruct *unstructured.Unstructured // Expected BOM
	var actualStruct *unstructured.Unstructured   // Generated BOM on given cluster

	json.Unmarshal(expected, &expectedStruct)
	json.Unmarshal(actual, &actualStruct)

	// Convert JSON to MAP for the container and application type
	CompareBOM(expectedStruct, actualStruct)
	return nil
}

func CompareBOM(expectedStruct *unstructured.Unstructured, actualStruct *unstructured.Unstructured) {
	var comparisonError = false
	// Extract BOM data to compare
	expectedContainerMap := ExtractBOMToMap(expectedStruct, CONTAINER)
	expectedApplicationMap := ExtractBOMToMap(expectedStruct, APPLICATION)

	actualContainerMap := ExtractBOMToMap(actualStruct, CONTAINER)
	actualApplicationMap := ExtractBOMToMap(actualStruct, APPLICATION)

	// Compare Golden and actual BOM, return response as Error and Warning
	containerMismatchWarning, containerMismatchError := CompareBOMData(expectedContainerMap, actualContainerMap)
	applicationMismatchWarning, applicationMismatchError := CompareBOMData(expectedApplicationMap, actualApplicationMap)

	// Print the error in table format
	if containerMismatchError != nil || applicationMismatchError != nil {
		fmt.Println("ERROR!")
		fmt.Println("Note: An error is when there is a mismatch in the actual cluster vs expected BOM")
		printMismatches(containerMismatchError, CONTAINER, false)
		printMismatches(applicationMismatchError, APPLICATION, false)

		comparisonError = true
	}

	// Print the warning in table format
	if containerMismatchWarning != nil || applicationMismatchWarning != nil {
		fmt.Println("\nWARNING!")
		fmt.Println("Note: A warning is when additional containers or applications are present in actual cluster or in expetced BOM")
		printMismatches(containerMismatchWarning, CONTAINER, true)
		printMismatches(applicationMismatchWarning, APPLICATION, true)
	}

	if comparisonError == true {
		fmt.Errorf("Error: Found mismatches between expetced BOM and actual cluster BOM")
	} else {
		fmt.Println("No errors found during BOM comparison.")
	}
}

func ExtractBOMToMap(data *unstructured.Unstructured, dataType string) map[string]map[string]string {
	dataMap := make(map[string]map[string]string)
	for _, value := range data.Object["components"].([]interface{}) {
		innerContainerMap := make(map[string]string)

		component := value.(map[string]interface{})
		bomType := component["type"].(string)
		name := component["name"].(string)
		purl := ecrRegex.ReplaceAllString(component["purl"].(string), "*")

		properties := component["properties"].([]interface{})
		if bomType == dataType { //Add to respective map for application or container type
			for _, props := range properties {
				itemKey := props.(map[string]interface{})["name"].(string)
				itemValue := props.(map[string]interface{})["value"].(string)
				innerContainerMap[itemKey] = itemValue
			}
			innerContainerMap[BOM_PROPERTY_NAME] = component["name"].(string)
			innerContainerMap[BOM_PROPERTY_VERSION] = component["version"].(string)
			if dataType == CONTAINER {
				purl = innerContainerMap[BOM_PROPERTY_CONTAINER_NAMESPACE] + "/" + ecrRegex.ReplaceAllString(name, "*")
			}
			dataMap[purl] = innerContainerMap
		}
	}

	return dataMap
}

func CompareBOMData(expected map[string]map[string]string, actual map[string]map[string]string) ([]ComponentData, []ComponentData) {
	var warnMismatching []ComponentData
	var errorMismatching []ComponentData

	// Check if the expected component from Expected BOM is present in the actual cluster BOM
	for name, expectedProps := range expected {
		actualProps, found := actual[name]

		friendlyName := expectedProps[BOM_PROPERTY_NAME]
		namespace := expectedProps[BOM_PROPERTY_CONTAINER_NAMESPACE]

		// If the component is not present in the actual cluster, we consider it as a warning
		if !found {
			comparisonProp := []ComparisonProperty{{PropertyName: "", Expected: "Exists",
				Actual: "Missing"}}
			mismatch := ComponentData{Name: /*name*/ namespace + "/" + friendlyName, Type: expectedProps[COMPONENT_TYPE], Properties: comparisonProp}
			warnMismatching = append(warnMismatching, mismatch)
		} else {
			propertyDiff := compareMaps(expectedProps, actualProps)
			// If the component is present in the actual cluster, but with different properties, we consider it as an error
			if propertyDiff != nil {
				mismatch := ComponentData{Name: /*name*/ namespace + "/" + friendlyName, Type: actualProps[COMPONENT_TYPE], Properties: propertyDiff}
				errorMismatching = append(errorMismatching, mismatch)
			}
		}
		// Compare the properties of the expected and actual components
	}

	// Check if the expected component from actual BOM is present in the expected cluster BOM
	for name, actualProps := range actual {
		_, found := expected[name]

		friendlyName := actualProps[BOM_PROPERTY_NAME]
		namespace := actualProps[BOM_PROPERTY_CONTAINER_NAMESPACE]

		// If the component is not present in the expected BOM, we consider it as a warning
		if !found {
			comparisonProp := []ComparisonProperty{{PropertyName: "", Expected: "Missing",
				Actual: "Exists"}}
			mismatch := ComponentData{Name: /*name*/ namespace + "/" + friendlyName, Type: actualProps[COMPONENT_TYPE], Properties: comparisonProp}
			warnMismatching = append(warnMismatching, mismatch)
		}
	}

	return warnMismatching, errorMismatching
}

func compareMaps(expected, actual map[string]string) []ComparisonProperty {
	var differences []ComparisonProperty

	for key, val1 := range expected {
		if val2, exists := actual[key]; !exists || ecrRegex.ReplaceAllString(val1, "*") != ecrRegex.ReplaceAllString(val2, "*") {
			comparisonProp := ComparisonProperty{PropertyName: key, Expected: val1, Actual: val2}
			differences = append(differences, comparisonProp)
		}
	}

	// Check for keys in actual that are not in expected
	for key, val2 := range actual {
		if _, exists := expected[key]; !exists {
			comparisonProp := ComparisonProperty{PropertyName: key, Expected: "Null", Actual: val2}
			differences = append(differences, comparisonProp)
		}
	}

	return differences
}

func printMismatches(errors []ComponentData, componentType string, skipPropertyName bool) {
	if len(errors) <= 0 {
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Skip setting 'PropertyName' column for BOM warning
	header := table.Row{"Type", "Name"}
	if !skipPropertyName {
		header = append(header, "Property Name")
	}
	header = append(header, getColumnName("BOM_EXPECTED_COLUMN_NAME", "Expected"), getColumnName("BOM_ACTUAL_COLUMN_NAME", "Actual"))
	t.AppendHeader(header)

	for _, differences := range errors {
		name, ctype := text.WrapHard(differences.Name, 50), componentType
		for _, values := range differences.Properties {
			row := table.Row{ctype, name}
			// Skip setting 'PropertyName' values for BOM warning
			if !skipPropertyName {
				row = append(row, values.PropertyName)
			}
			row = append(row, text.WrapHard(getColoredText(values.Expected, green), 50), text.WrapHard(getColoredText(values.Actual, red), 50))
			t.AppendRows([]table.Row{row})
			name, ctype = "", ""
		}
		t.AppendSeparator()
	}
	t.Render()
}

func getColoredText(text string, color string) string {
	// Override the default color for BOM warning messages
	if strings.Contains(text, "Missing") {
		color = red
	} else if strings.Contains(text, "Exists") {
		color = green
	}
	return fmt.Sprint(color, text, reset)
}
func getColumnName(envClusterName string, defaultName string) (clusterName string) {
	clusterName = os.Getenv(envClusterName)
	if clusterName == "" {
		return defaultName
	}
	return clusterName
}
