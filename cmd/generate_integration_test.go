package cmd_test

import (
	. "cluster-codex/cmd"
	"cluster-codex/internal/k8"
	"cluster-codex/internal/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GenerateBOM - Integration", Label("integration"), Ordered, func() {
	const testNamespace = "clx-test"
	const testNamespace2 = "clx-test-2"
	var k8client *k8.K8sClient
	var bom *model.BOM
	var err error

	Context("When generate BOM is called should return valid BOM", func() {
		BeforeEach(func() {
			k8client, err = k8.GetClient()
			if err != nil {
				Fail(err.Error())
			}
		})

		DescribeTable("should have valid metadata, components and images",
			func(namespaces []string, findNamespace bool, findNamespace2 bool) {
				// Set the filter for the namespaces
				k8.K8Filter = model.Filter{NamespacedInclusions: []model.NamespacedInclusion{{Namespaces: namespaces}}}
				InitializeFilterStruct(&k8.K8Filter)

				bom, err = GenerateBOM(k8client)

				Expect(err).To(BeNil())
				Expect(bom).ToNot(BeNil())
				Expect(bom.BomFormat).To(Equal("CycloneDX"))
				Expect(bom.SpecVersion).To(Equal("1.6"))
				Expect(bom.Metadata).To(Not(BeNil()))
				Expect(len(bom.Components)).To(BeNumerically(">", 0))

				// If we filter out the testNamespace, we will not capture any Deployments in it.
				components := bom.FindApplications("nginx-deployment", "Deployment", testNamespace)
				Expect(len(components)).To(BeNumerically("==", map[bool]int{true: 1, false: 0}[findNamespace]))
				if len(components) > 0 {
					Expect(components[0].PackageURL).To(Equal("pkg:k8s/Deployment/nginx-deployment?apiVersion=apps%2Fv1&namespace=clx-test"))
				}

				// Even when we are filtering out the test namespace, it will be in the BOM since it is not in a namespace.
				components = bom.FindApplications(testNamespace, "Namespace", "")
				Expect(len(components)).To(BeNumerically("==", 1))

				// Non-existent pods never appear.
				components = bom.FindApplications("non-existent-pod", "Pod", "default")
				Expect(len(components)).To(BeNumerically("==", 0))

				// check images
				images := bom.FindContainers("index.docker.io/library/nginx", "Image", testNamespace)
				if findNamespace {
					Expect(len(images)).To(BeNumerically("==", 1))
					ownerRef, found := images[0].GetProperty("clx:k8s:ownerRef")
					Expect(found).To(BeTrue())
					Expect(ownerRef).To(Equal("Deployment/nginx-deployment"))
					//Image sha will be different for multi-arch images so checking substring
					Expect(images[0].PackageURL).To(ContainSubstring("pkg:oci/library/nginx@sha256:"))
					Expect(images[0].PackageURL).To(ContainSubstring("?namespace=clx-test&ownerRef=Deployment%2Fnginx-deployment&repository_url=index.docker.io%2Flibrary%2Fnginx"))
				} else {
					Expect(len(images)).To(BeNumerically("==", 0))
				}

				// Check if the BOM contains the components from the other namespace.
				components1 := bom.FindApplications("nginx-deployment", "Deployment", testNamespace2)
				Expect(len(components1)).To(BeNumerically("==", map[bool]int{true: 1, false: 0}[findNamespace2]))
				if len(components1) > 0 {
					Expect(components1[0].PackageURL).To(Equal("pkg:k8s/Deployment/nginx-deployment?apiVersion=apps%2Fv1&namespace=clx-test-2"))
				}

				// check images
				images1 := bom.FindContainers("index.docker.io/library/nginx", "Image", testNamespace2)
				if findNamespace2 {
					Expect(len(images1)).To(BeNumerically("==", 1))
					ownerRef, found := images1[0].GetProperty("clx:k8s:ownerRef")
					Expect(found).To(BeTrue())
					Expect(ownerRef).To(Equal("Deployment/nginx-deployment"))
					//Image sha will be different for multi-arch images so checking substring
					Expect(images1[0].PackageURL).To(ContainSubstring("pkg:oci/library/nginx@sha256:"))
					Expect(images1[0].PackageURL).To(ContainSubstring("?namespace=clx-test-2&ownerRef=Deployment%2Fnginx-deployment&repository_url=index.docker.io%2Flibrary%2Fnginx"))
				} else {
					Expect(len(images1)).To(BeNumerically("==", 0))
				}

			},
			Entry("No Namespaces", []string{}, true, true),
			Entry("Specific Test Namespace", []string{testNamespace}, true, false),
			Entry("Multiple valid test namespaces", []string{testNamespace, testNamespace2}, true, true),
			Entry("Specific Test Namespace and another non existent one", []string{testNamespace, "banana-lemon"}, true, false),
			Entry("Existing Namespace without image", []string{"default"}, false, false),
		)
	})
})
