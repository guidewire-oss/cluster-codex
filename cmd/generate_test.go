package cmd_test

import (
	. "cluster-codex/cmd"
	"cluster-codex/internal/k8/k8fakes"
	"cluster-codex/internal/model"
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("constructPath", Label("unit"), func() {
	Context("when given a valid relative path", func() {
		It("should return nil", func() {
			filePath := "user/folder/example.txt"
			err := ValidatePath(filePath)
			Expect(err).To(BeNil())
		})
	})

	Context("when given a path with invalid characters", func() {
		It("should return an error with the invalid character", func() {
			filePath := "invalid|path/example.txt"
			err := ValidatePath(filePath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("path contains invalid character: |"))
		})
	})

	Context("when given an empty path", func() {
		It("should return an error", func() {
			filePath := ""
			err := ValidatePath(filePath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("path cannot be empty"))
		})
	})

	Context("when given a path without a file extension", func() {
		It("should return an error", func() {
			filePath := "user/folder/example"
			err := ValidatePath(filePath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("path must have a file extension"))
		})
	})
})

var _ = Describe("GenerateBOM - Unit", Label("unit"), func() {
	var fakeK8sClient *k8fakes.FakeK8sClientInterface

	BeforeEach(func() {
		fakeK8sClient = new(k8fakes.FakeK8sClientInterface)
	})

	Context("when GetAllComponents returns components", func() {
		It("should return a BOM with components", func() {
			mockComponent := model.Component{
				Type:       "application",
				Name:       "test-component",
				Version:    "1.0.0",
				PackageURL: "pkg:docker/test-component@1.0.0", // Optional (omit if not needed)
				Properties: []model.Property{
					{Name: model.ComponentKind, Values: []string{"HelmChart"}},
					{Name: model.ComponentNamespace, Values: []string{"flux-system"}},
					{Name: model.ComponentVersion, Values: []string{"1.0.0"}},
				},
			}
			mockResponse := []model.Component{mockComponent}
			fakeK8sClient.GetAllComponentsStub = func(ctx context.Context) ([]model.Component, error) {
				return mockResponse, nil
			}

			bom := GenerateBOM(fakeK8sClient, []string{})

			Expect(bom).ToNot(BeNil())
			Expect(bom.Components).To(HaveLen(1))
			Expect(bom.Components[0].Name).To(Equal("test-component"))
		})
	})

	Context("when GetAllComponents returns an error", func() {
		It("should return nil BOM", func() {
			fakeK8sClient.GetAllComponentsStub = func(ctx context.Context) ([]model.Component, error) {
				return nil, assert.AnError
			}

			bom := GenerateBOM(fakeK8sClient, []string{})

			Expect(bom).To(BeNil())
		})
	})
})
