package model_test

import (
	"cluster-codex/internal/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type TestFilterFields struct {
	NonNamespacedInclusions model.NonNamespacedInclusions
	NamespacedInclusions    []model.NamespacedInclusion
}

var _ = Describe("Filter", Label("unit"), func() {
	DescribeTable("ShouldIncludeThisResource",
		func(f TestFilterFields, resourceNamespace string, resourceKind string, want bool) {
			filter := &model.Filter{
				NonNamespacedInclusions: f.NonNamespacedInclusions,
				NamespacedInclusions:    f.NamespacedInclusions,
			}
			shouldInclude := filter.ShouldIncludeThisResource(resourceNamespace, resourceKind)
			Expect(shouldInclude).To(Equal(want))
		},
		Entry("should return true when all namespaces are included", TestFilterFields{
			NonNamespacedInclusions: model.NonNamespacedInclusions{},
			NamespacedInclusions:    []model.NamespacedInclusion{},
		}, "test-ns", "Pod", true),
		Entry("should return true when a namespace is included", TestFilterFields{
			NonNamespacedInclusions: model.NonNamespacedInclusions{},
			NamespacedInclusions:    []model.NamespacedInclusion{{Namespaces: []string{"test-ns"}}},
		}, "test-ns", "Pod", true),
		Entry("should return true when the namespace is *", TestFilterFields{
			NonNamespacedInclusions: model.NonNamespacedInclusions{},
			NamespacedInclusions:    []model.NamespacedInclusion{{Namespaces: []string{"*"}}},
		}, "test-ns", "Pod", true),
		Entry("should return false when the namespace is not included", TestFilterFields{
			NonNamespacedInclusions: model.NonNamespacedInclusions{},
			NamespacedInclusions:    []model.NamespacedInclusion{{Namespaces: []string{"kube-system"}}},
		}, "test-ns", "Pod", false),
		Entry("should return true when the resource is included", TestFilterFields{
			NonNamespacedInclusions: model.NonNamespacedInclusions{},
			NamespacedInclusions:    []model.NamespacedInclusion{{Namespaces: []string{"*"}, Resources: []string{"Pod"}}},
		}, "test-ns", "Pod", true),
		Entry("should return true when the resource is *", TestFilterFields{
			NonNamespacedInclusions: model.NonNamespacedInclusions{},
			NamespacedInclusions:    []model.NamespacedInclusion{{Namespaces: []string{"*"}, Resources: []string{"*"}}},
		}, "test-ns", "Pod", true),
		Entry("should return false when the resource is not included", TestFilterFields{
			NonNamespacedInclusions: model.NonNamespacedInclusions{},
			NamespacedInclusions:    []model.NamespacedInclusion{{Namespaces: []string{"*"}, Resources: []string{"Deployment"}}},
		}, "test-ns", "Pod", false),
		Entry("should return false when the resource is included but namespace is not included", TestFilterFields{
			NonNamespacedInclusions: model.NonNamespacedInclusions{},
			NamespacedInclusions:    []model.NamespacedInclusion{{Namespaces: []string{"kube-system", "default"}, Resources: []string{"Deployment"}}},
		}, "test-ns", "Pod", false),
	)

	DescribeTable("IncludesAllKindsNonNamespaced", Label("unit"),
		func(resources []string, want bool) {
			filter := &model.Filter{
				NonNamespacedInclusions: model.NonNamespacedInclusions{
					Resources: resources,
				},
			}
			Expect(filter.IncludesAllKindsNonNamespaced()).To(Equal(want))
		},
		Entry("should return true when resources list is empty", []string{}, true),
		Entry("should return true when resources list contains '*'", []string{"*"}, true),
		Entry("should return false when resources list contains specific resources", []string{"pod", "service"}, false),
	)
})
