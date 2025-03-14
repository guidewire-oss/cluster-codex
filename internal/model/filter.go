package model

import "cluster-codex/internal/utils"

type Filter struct {
	NonNamespacedInclusions NonNamespacedInclusions `json:"non-namespaced-inclusions"`
	NamespacedInclusions    []NamespacedInclusion   `json:"namespaced-inclusions"`
}

// Inclusion - Struct to match JSON structure
type NamespacedInclusion struct {
	Namespaces []string `json:"namespaces"`
	Resources  []string `json:"resources"`
}

type NonNamespacedInclusions struct {
	Resources []string `json:"resources"`
}

func (filter *Filter) ShouldIncludeThisResource(namespace string, kind string) bool {
	if len(filter.NamespacedInclusions) == 0 {
		return true
	}
	for _, f := range filter.NamespacedInclusions {
		if f.Namespaces == nil || f.Namespaces[0] == "*" || utils.Contains(f.Namespaces, namespace) {
			if len(f.Resources) == 0 || f.Resources[0] == "*" || utils.Contains(f.Resources, kind) {
				return true
			}
		}
	}
	return false
}

func (filter *Filter) GetNamespaceList() []string {

	var namespaces []string
	for _, f := range filter.NamespacedInclusions {
		namespaces = append(namespaces, f.Namespaces...)
	}
	// If any of the namespaceInclusion filter contains a *, then return empty list so that we get images from all namespaces
	if utils.Contains(namespaces, "*") {
		return []string{}
	}
	return namespaces
}

// IncludesAllKindsNonNamespaced checks if the filter includes all non-namespaced resources.
//
// The function returns true if the filter does not specify any non-namespaced resources or if it includes all non-namespaced resources by using the wildcard "*".
// Otherwise, it returns false.

func (filter *Filter) IncludesAllKindsNonNamespaced() bool {
	f := filter.NonNamespacedInclusions
	if len(f.Resources) == 0 || f.Resources[0] == "*" {
		return true
	}
	return false
}
