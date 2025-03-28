package model

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"sort"
	"time"
)

// CustomTime enforces RFC 3339 without milliseconds
type CustomTime time.Time

// RFC 3339 format without milliseconds
const timeFormat = "2006-01-02T15:04:05Z07:00"

const ComponentKind = "clx:k8s:componentKind"
const ComponentNamespace = "clx:k8s:componentNamespace"
const ComponentVersion = "clx:k8s:componentVersion"
const ComponentOwnerRef = "clx:k8s:ownerRef"
const ComponentSourceRef = "clx:k8s:source"

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
	existingProperty, found := component.GetPropertyObject(key)
	if found {
		existingProperty.InsertValue(value)
	} else {
		prop := Property{
			Name:   key,
			Values: []string{value},
		}
		component.Properties = append(component.Properties, prop)
	}
}

func (component *Component) AddPropertyMultipleValue(key string, value ...string) {
	existingProperty, found := component.GetPropertyObject(key)
	if found {
		existingProperty.InsertValue(value...)
	} else {
		prop := Property{
			Name:   key,
			Values: value,
		}
		component.Properties = append(component.Properties, prop)
	}
}

func (component *Component) GetPropertyObject(name string) (*Property, bool) {
	// If we do "for _, prop := range component.Properties" then prop is a copy. So do it like this.
	for i := range component.Properties {
		if component.Properties[i].Name == name {
			return &component.Properties[i], true
		}
	}
	return nil, false
}

func (component *Component) GetProperty(name string) (string, bool) {
	existingProperty, found := component.GetPropertyObject(name)
	if found {
		return existingProperty.Values[0], true
	}
	return "", false
}

func (component *Component) GetKind() string {
	value, found := component.GetProperty(ComponentKind)
	if found {
		return value
	}
	return ""
}

func (component *Component) GetNamespace() string {
	value, found := component.GetProperty(ComponentNamespace)
	if found {
		return value
	}
	return ""
}

// ByComponentSorting Sorting implementation for Components
type ByComponentSorting []Component

func (c ByComponentSorting) Len() int { return len(c) }
func (c ByComponentSorting) Less(i, j int) bool {
	// 1️⃣ Sort by Type: "application" < "container"
	typeOrder := map[string]int{"application": 0, "container": 1}
	ti, tj := typeOrder[c[i].Type], typeOrder[c[j].Type]
	if ti != tj {
		return ti < tj
	}

	// 2️⃣ Sort by `clx:k8s:componentKind`
	kindI, kindJ := c[i].GetKind(), c[j].GetKind()
	if kindI != kindJ {
		return kindI < kindJ
	}

	// 3️⃣ Sort by Name
	if c[i].Name != c[j].Name {
		return c[i].Name < c[j].Name
	}
	// 4️⃣ Sort by Namespace
	return c[i].GetNamespace() < c[j].GetNamespace()
}

func (c ByComponentSorting) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (bom *BOM) Sort() {
	// Sort the components
	sort.Sort(ByComponentSorting(bom.Components))

	// Sort the properties in each component by name
	for i := range bom.Components {
		sort.Sort(ByPropertyName(bom.Components[i].Properties))
	}
}

func (bom *BOM) FindApplications(name string, kind string, namespace string) []Component {
	return bom.findComponents("application", name, kind, namespace)
}

func (bom *BOM) FindContainers(name string, kind string, namespace string) []Component {
	return bom.findComponents("container", name, kind, namespace)
}

func (bom *BOM) findComponents(componentType string, name string, kind string, namespace string) []Component {
	var returnComponents []Component = make([]Component, 0)
	for _, component := range bom.Components {
		if componentType == "" || component.Type == componentType {
			if name == "" || component.Name == name {
				props, found := component.GetProperty(ComponentKind)
				if found && props == kind {
					if namespace != "" {
						props, found := component.GetProperty(ComponentNamespace)
						if found && props == namespace {
							returnComponents = append(returnComponents, component)
						}
						continue
					}
					returnComponents = append(returnComponents, component)
				}
			}
		}
	}
	return returnComponents
}

func (bom *BOM) FindApplicationsByKind(kind string, namespace string) []Component {
	return bom.findComponentsByKind("application", kind, namespace)
}

func (bom *BOM) FindContainersByKind(kind string, namespace string) []Component {
	return bom.findComponentsByKind("container", kind, namespace)
}

func (bom *BOM) findComponentsByKind(componentType string, kind string, namespace string) []Component {
	var returnComponents []Component = make([]Component, 0)
	for _, component := range bom.Components {
		if componentType == "" || component.Type == componentType {
			props, found := component.GetProperty(ComponentKind)
			if found && props == kind {
				if namespace != "" {
					props, found := component.GetProperty(ComponentNamespace)
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

func (p Property) MarshalJSON() ([]byte, error) {
	if len(p.Values) == 1 {
		return json.Marshal(struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}{p.Name, p.Values[0]})
	}
	return json.Marshal(struct {
		Name   string   `json:"name"`
		Values []string `json:"value"`
	}{p.Name, p.Values})
}

// UnmarshalJSON customizes JSON decoding: supports single string or array and sorts values.
func (p *Property) UnmarshalJSON(data []byte) error {
	var single struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(data, &single); err == nil {
		p.Name = single.Name
		p.Values = []string{single.Value}
		return nil
	}

	var multi struct {
		Name   string   `json:"name"`
		Values []string `json:"value"`
	}
	if err := json.Unmarshal(data, &multi); err == nil {
		p.Name = multi.Name
		p.Values = multi.Values
		sort.Strings(p.Values) // ✅ Ensure values are always sorted after unmarshaling
		return nil
	}

	return fmt.Errorf("invalid property format")
}

// Property represents a component property that can have one or multiple values.
type Property struct {
	Name   string   `json:"name"`
	Values []string `json:"value"` // Use "value" in JSON for both single and multiple values
}

// ByPropertyName Sorting implementation for Properties (by Name)
type ByPropertyName []Property

func (p ByPropertyName) Len() int           { return len(p) }
func (p ByPropertyName) Less(i, j int) bool { return p[i].Name < p[j].Name } // Sort ascending
func (p ByPropertyName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p *Property) InsertValue(values ...string) {
	sort.Strings(values)
	for _, value := range values {
		// Find the index where the value should be inserted.
		index := sort.SearchStrings(p.Values, value)

		// If the value is already present, do nothing (optional)
		if index < len(p.Values) && p.Values[index] == value {
			return
		}

		// Insert the value at the correct position
		p.Values = append(p.Values, "")
		copy(p.Values[index+1:], p.Values[index:])
		p.Values[index] = value
	}
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
