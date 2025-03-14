package cmd_test

import (
	. "cluster-codex/cmd"
	"cluster-codex/internal/k8"
	"cluster-codex/internal/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GenerateBOM - Integration", Label("integration"), func() {
	var k8client *k8.K8sClient
	var bom *model.BOM
	var err error

	Context("When generate BOM is called should return valid BOM", func() {
		BeforeEach(func() {
			k8client, err = k8.GetClient()
			if err != nil {
				Fail(err.Error())
			}
			bom = GenerateBOM(k8client)
		})

		It("should have valid metadata, components and images", func() {
			Expect(bom).ToNot(BeNil())
			Expect(bom.BomFormat).To(Equal("CycloneDX"))
			Expect(bom.SpecVersion).To(Equal("1.6"))
			Expect(bom.Metadata).To(Not(BeNil()))
			Expect(len(bom.Components)).To(BeNumerically(">", 0))

			components := bom.FindApplications("nginx-deployment", "Deployment", "test")
			Expect(len(components)).To(BeNumerically("==", 1))
			Expect(components[0].PackageURL).To(Equal("pkg:k8s/Deployment/nginx-deployment?apiVersion=apps%2Fv1&namespace=test"))

			components = bom.FindApplications("test", "Namespace", "")
			Expect(len(components)).To(BeNumerically("==", 1))
			components = bom.FindApplications("non-existent-pod", "Pod", "default")
			Expect(len(components)).To(BeNumerically("==", 0))

			// check images
			images := bom.FindContainers("index.docker.io/library/nginx", "Image", "test")
			Expect(len(images)).To(BeNumerically("==", 1))
			ownerRef, found := images[0].GetProperty("clx:k8s:ownerRef")
			Expect(found).To(BeTrue())
			Expect(ownerRef).To(Equal("Deployment/nginx-deployment"))
		})

	})
})
