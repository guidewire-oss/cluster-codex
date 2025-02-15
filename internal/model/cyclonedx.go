package model

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

// CustomTime enforces RFC 3339 without milliseconds
type CustomTime time.Time

// RFC 3339 format without milliseconds
const timeFormat = "2006-01-02T15:04:05Z07:00"

// MarshalJSON formats time correctly
func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", time.Time(*ct).Format(timeFormat))
	return []byte(formatted), nil
}

// UnmarshalJSON parses the time correctly
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	t, err := time.Parse(`"`+timeFormat+`"`, string(data))
	if err != nil {
		return err
	}
	*ct = CustomTime(t)
	return nil
}

// NewBOM creates a BOM with a valid RFC 4122 UUID as SerialNumber
func NewBOM() *BOM {
	creationTime := CustomTime(time.Now())
	return &BOM{
		BomFormat:    "CycloneDX",
		SpecVersion:  "1.6",
		SerialNumber: uuid.New().String(), // Generate a valid RFC 4122 UUID
		Version:      1,
		Metadata: &Metadata{
			Timestamp: &creationTime,
			Tools: []Tool{
				Tool{
					Vendor:  clusterCodexVendor,
					Name:    clusterCodex,
					Version: clusterCodexVersion,
				},
			},
			Component: &Component{
				Name:    "kubernetes",
				Version: "", // This should be populated after the server connection
				Type:    "platform",
			},
		},
	}
}

// BOM represents the CycloneDX Bill of Materials
type BOM struct {
	BomFormat    string      `json:"bomFormat"`
	SpecVersion  string      `json:"specVersion"`
	SerialNumber string      `json:"serialNumber,omitempty"`
	Version      int         `json:"version"`
	Metadata     *Metadata   `json:"metadata,omitempty"`
	Components   []Component `json:"components,omitempty"`
}

// Metadata provides information about the SBOM creation
type Metadata struct {
	Timestamp *CustomTime `json:"timestamp"`
	Tools     []Tool      `json:"tools"`
	Component *Component  `json:"component"`
}

// Tool represents the software that generated the SBOM
type Tool struct {
	Vendor  string `json:"vendor"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Component defines a software package or library
type Component struct {
	Type       string     `json:"type"`
	Name       string     `json:"name"`
	Version    string     `json:"version"`
	PackageURL string     `json:"purl,omitempty"`
	Properties []Property `json:"properties,omitempty"`
	Licenses   []License  `json:"licenses,omitempty"`
	Hashes     []Hash     `json:"hashes,omitempty"`
}

func (component *Component) AddProperty(key string, value string) {
	prop := Property{
		Name:  key,
		Value: value,
	}
	component.Properties = append(component.Properties, prop)
}

func (component *Component) GetProperty(name string) (string, bool) {
	for _, prop := range component.Properties {
		if prop.Name == name {
			return prop.Value, true
		}
	}
	return "", false
}

func (bom *BOM) FindComponents(name string, kind string, namespace string) []Component {
	var returnComponents []Component = make([]Component, 0)
	for _, component := range bom.Components {
		if component.Name == name {
			props, found := component.GetProperty("clx:k8s:componentKind")
			if found && props == kind {
				if namespace != "" {
					props, found := component.GetProperty("clx:k8s:namespace")
					if found && props == namespace {
						returnComponents = append(returnComponents, component)
					}
					continue
				}
				returnComponents = append(returnComponents, component)
			}
		}
	}
	return returnComponents
}

func (bom *BOM) FindComponentsByKind(kind string, namespace string) []Component {
	var returnComponents []Component = make([]Component, 0)
	for _, component := range bom.Components {
		props, found := component.GetProperty("clx:k8s:componentKind")
		if found && props == kind {
			if namespace != "" {
				props, found := component.GetProperty("clx:k8s:namespace")
				if found && props == namespace {
					returnComponents = append(returnComponents, component)
				}
				continue
			}
			returnComponents = append(returnComponents, component)
		}
	}
	return returnComponents
}

type Property struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// License represents licensing information
type License struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Hash represents cryptographic hashes for verification
type Hash struct {
	Algorithm string `json:"alg"`
	Value     string `json:"value"`
}
