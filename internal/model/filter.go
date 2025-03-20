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

// This function is used to determine whether the filter includes all namespaces. If filter does not specify namespaces, it returns true.
func (filter *Filter) IncludesAllNamespaces() bool {
	for _, f := range filter.NamespacedInclusions {
		if f.Namespaces == nil || len(f.Namespaces) == 0 || f.Namespaces[0] == "*" {
			return true
		}
	}
	return false
}

func (filter *Filter) ShouldIncludesThisResource(namespace string, kind string) bool {
	for _, f := range filter.NamespacedInclusions {
		if f.Namespaces == nil || f.Namespaces[0] == "*" || utils.Contains(f.Namespaces, namespace) {
			if utils.Contains(f.Resources, kind) || f.Resources[0] == "*" {
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

//func (filter *Filter) IsKindInGivenNamespace() []string {
//
//}
