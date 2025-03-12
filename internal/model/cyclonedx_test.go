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

var _ = Describe("StaticCycloneEntityToJson - Unit", Label("unit"), func() {
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

var _ = Describe("StaticCycloneEntitySorting - Unit", Label("unit"), func() {
	It("should correctly sort the components by property name", func() {
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
				{Type: "application", Name: "My Application Sorted", Version: "1.0.0", PackageURL: "pkg:maven/com.test/lib@1.0.0",
					Properties: []Property{
						{Name: ComponentKind, Values: []string{"ServiceAccount"}},
						{Name: ComponentNamespace, Values: []string{"kube-system"}},
					},
				},
				{Type: "application", Name: "My Application Unsorted", Version: "1.0.0", PackageURL: "pkg:maven/com.test/lib@1.0.0",
					Properties: []Property{
						{Name: ComponentNamespace, Values: []string{"kube-system"}},
						{Name: ComponentKind, Values: []string{"ServiceAccount"}},
					},
				},
			},
		}

		// Sort the BOM
		testBOM.Sort()
		// ✅ Assertion that the BOM component properties are sorted
		Expect(len(testBOM.Components)).To(Equal(2))
		Expect(testBOM.Components[0].Name).To(Equal("My Application Sorted"))
		Expect(testBOM.Components[1].Name).To(Equal("My Application Unsorted"))

		// Assert the properties
		Expect(len(testBOM.Components[0].Properties)).To(Equal(2))
		Expect(testBOM.Components[0].Properties[0].Name).To(Equal(ComponentKind))
		Expect(testBOM.Components[0].Properties[0].Values[0]).To(Equal("ServiceAccount"))
		Expect(testBOM.Components[0].Properties[1].Name).To(Equal(ComponentNamespace))
		Expect(testBOM.Components[0].Properties[1].Values[0]).To(Equal("kube-system"))
		Expect(len(testBOM.Components[1].Properties)).To(Equal(2))
		Expect(testBOM.Components[1].Properties[0].Name).To(Equal(ComponentKind))
		Expect(testBOM.Components[1].Properties[0].Values[0]).To(Equal("ServiceAccount"))
		Expect(testBOM.Components[1].Properties[1].Name).To(Equal(ComponentNamespace))
		Expect(testBOM.Components[1].Properties[1].Values[0]).To(Equal("kube-system"))
	})
	It("should correctly sort the components by type, then kind, then name, then namespace", func() {
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
				{Type: "application", Name: "My Application A ServiceAccount 2", Version: "1.0.0", PackageURL: "pkg:maven/com.test/lib@1.0.0",
					Properties: []Property{
						{Name: ComponentKind, Values: []string{"ServiceAccount"}},
						{Name: ComponentNamespace, Values: []string{"kube-system-3"}},
					},
				},
				{Type: "container", Name: "My Container 2", Version: "1.0.0", PackageURL: "pkg:maven/com.test/lib@1.0.0",
					Properties: []Property{
						{Name: ComponentKind, Values: []string{"Image"}},
						{Name: ComponentNamespace, Values: []string{"kube-system"}},
					},
				},
				{Type: "application", Name: "My Application A ServiceAccount 2", Version: "1.0.0", PackageURL: "pkg:maven/com.test/lib@1.0.0",
					Properties: []Property{
						{Name: ComponentKind, Values: []string{"ServiceAccount"}},
						{Name: ComponentNamespace, Values: []string{"kube-system-1"}},
					},
				},
				{Type: "application", Name: "My Application A ServiceAccount 2", Version: "1.0.0", PackageURL: "pkg:maven/com.test/lib@1.0.0",
					Properties: []Property{
						{Name: ComponentKind, Values: []string{"ServiceAccount"}},
						{Name: ComponentNamespace, Values: []string{"kube-system-2"}},
					},
				},
				{Type: "application", Name: "My Application A ServiceAccount 1", Version: "1.0.0", PackageURL: "pkg:maven/com.test/lib@1.0.0",
					Properties: []Property{
						{Name: ComponentKind, Values: []string{"ServiceAccount"}},
						{Name: ComponentNamespace, Values: []string{"kube-system"}},
					},
				},
				{Type: "application", Name: "My Application Namespace", Version: "1.0.0", PackageURL: "pkg:maven/com.test/lib@1.0.0",
					Properties: []Property{
						{Name: ComponentKind, Values: []string{"Namespace"}},
						{Name: ComponentNamespace, Values: []string{""}},
					},
				},
				{Type: "container", Name: "My Container 1", Version: "1.0.0", PackageURL: "pkg:maven/com.test/lib@1.0.0",
					Properties: []Property{
						{Name: ComponentNamespace, Values: []string{"kube-system"}},
						{Name: ComponentKind, Values: []string{"Image"}},
					},
				},
			},
		}

		// Sort the BOM
		testBOM.Sort()
		// ✅ Assertion that the BOM component properties are sorted
		Expect(len(testBOM.Components)).To(Equal(7))
		Expect(testBOM.Components[0].Name).To(Equal("My Application Namespace"))
		Expect(testBOM.Components[0].GetKind()).To(Equal("Namespace"))
		Expect(testBOM.Components[1].Name).To(Equal("My Application A ServiceAccount 1"))
		Expect(testBOM.Components[1].GetKind()).To(Equal("ServiceAccount"))
		Expect(testBOM.Components[1].GetNamespace()).To(Equal("kube-system"))
		Expect(testBOM.Components[2].Name).To(Equal("My Application A ServiceAccount 2"))
		Expect(testBOM.Components[2].GetKind()).To(Equal("ServiceAccount"))
		Expect(testBOM.Components[2].GetNamespace()).To(Equal("kube-system-1"))
		Expect(testBOM.Components[3].Name).To(Equal("My Application A ServiceAccount 2"))
		Expect(testBOM.Components[3].GetKind()).To(Equal("ServiceAccount"))
		Expect(testBOM.Components[3].GetNamespace()).To(Equal("kube-system-2"))
		Expect(testBOM.Components[4].Name).To(Equal("My Application A ServiceAccount 2"))
		Expect(testBOM.Components[4].GetKind()).To(Equal("ServiceAccount"))
		Expect(testBOM.Components[4].GetNamespace()).To(Equal("kube-system-3"))
		Expect(testBOM.Components[5].Name).To(Equal("My Container 1"))
		Expect(testBOM.Components[6].Name).To(Equal("My Container 2"))
	})
})
