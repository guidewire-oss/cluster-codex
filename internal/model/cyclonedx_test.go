package model_test

import (
	. "cluster-codex/internal/model"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("StaticCycloneEntityToJson", func() {
	It("should correctly marshal BOM to JSON", func() {
		// Fixed timestamp: January 31, 2025, at 12:00:00 PST
		pst := time.FixedZone("PST", -8*3600)
		timestamp := CustomTime(time.Date(2025, time.January, 31, 12, 0, 0, 223344, pst))

		// Define a BOM object
		testBOM := BOM{
			BomFormat:    "CycloneDX",
			SpecVersion:  "1.6",
			SerialNumber: uuid.New().String(),
			Version:      1,
			Metadata: &Metadata{
				Timestamp: &timestamp,
				Tools: []Tool{
					{Vendor: "ToolVendor", Name: "SBOMGenerator", Version: "1.0.0"},
				},
				Component: &Component{
					Type: "library", Name: "example-lib", Version: "1.0.0", PackageURL: "pkg:maven/com.example/lib@1.0.0",
					Properties: []Property{
						{Name: "componentType", Values: []string{"cluster"}},
						{Name: "componentName", Values: []string{"mint"}},
					},
				},
			},
			Components: []Component{
				{Type: "library", Name: "example-lib", Version: "1.0.0", PackageURL: "pkg:maven/com.example/lib@1.0.0"},
			},
		}

		// Marshal BOM to JSON
		jsonOutput, err := json.MarshalIndent(testBOM, "", "  ")

		// ✅ Assertions with Gomega
		Expect(err).ToNot(HaveOccurred(), "Failed to marshal BOM")
		Expect(string(jsonOutput)).To(ContainSubstring(`"timestamp": "2025-01-31T12:00:00-08:00"`), "Timestamp format incorrect")
		Expect(string(jsonOutput)).To(ContainSubstring(`"bomFormat": "CycloneDX"`), "BOM format missing")
		Expect(string(jsonOutput)).To(ContainSubstring(`"specVersion": "1.6"`), "Spec version missing")
		Expect(string(jsonOutput)).To(ContainSubstring(`"name": "componentType"`), "Property componentType missing")
		Expect(string(jsonOutput)).To(ContainSubstring(`"components"`), "Components missing")

		// Print JSON output (optional)
		fmt.Println(string(jsonOutput))
	})
})
