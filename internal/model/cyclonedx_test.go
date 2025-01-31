package model

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestStaticCycloneEntityToJson(t *testing.T) {
	// Fixed timestamp: January 31, 2025, at 12:00:00 PST. Add some additional time to show dropped by formatting.
	// Define Pacific Standard Time (PST) timezone (UTC-08:00)
	pst := time.FixedZone("PST", -8*3600)
	timestamp := CustomTime(time.Date(2025, time.January, 31, 12, 0, 0, 223344, pst))
	bom := BOM{
		BomFormat:    "CycloneDX",
		SpecVersion:  "1.6",
		SerialNumber: uuid.New().String(),
		Version:      1,
		Metadata: &Metadata{
			Timestamp: &timestamp,
			Tools: []Tool{
				{Vendor: "ToolVendor", Name: "SBOMGenerator", Version: "1.0.0"},
			},
			Component: &Component{Type: "library", Name: "example-lib", Version: "1.0.0", PackageURL: "pkg:maven/com.example/lib@1.0.0",
				Properties: []Property{
					{Name: "componentType", Value: "cluster"},
					{Name: "componentName", Value: "mint"},
				},
			},
		},
		Components: []Component{
			{Type: "library", Name: "example-lib", Version: "1.0.0", PackageURL: "pkg:maven/com.example/lib@1.0.0"},
		},
	}

	jsonOutput, err := json.MarshalIndent(bom, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
	}
	assert.NoError(t, err, "Failed to marshal BOM")
	// Check some specific values in the output.
	assert.Contains(t, string(jsonOutput), `"timestamp": "2025-01-31T12:00:00-08:00"`, "Timestamp format incorrect")
	assert.Contains(t, string(jsonOutput), `"bomFormat": "CycloneDX"`, "BOM format missing")
	assert.Contains(t, string(jsonOutput), `"specVersion": "1.6"`, "Spec version missing")
	assert.Contains(t, string(jsonOutput), `"name": "componentType"`, "Property componentType missing")
	assert.Contains(t, string(jsonOutput), `"components"`, "Components missing")
	// Print the JSON output
	fmt.Println(string(jsonOutput))
}
