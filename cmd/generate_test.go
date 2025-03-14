package cmd_test

import (
	. "cluster-codex/cmd"
	"cluster-codex/internal/k8/k8fakes"
	"cluster-codex/internal/model"
	"context"
	"encoding/json"
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
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
			fakeK8sClient.GetAllComponentsStub = func(ctx context.Context) ([]model.Component, []string, error) {
				return mockResponse, []string{}, nil
			}

			bom, err := GenerateBOM(fakeK8sClient)

			Expect(err).To(BeNil())
			Expect(bom).ToNot(BeNil())
			Expect(bom.Components).To(HaveLen(1))
			Expect(bom.Components[0].Name).To(Equal("test-component"))
		})
	})

	Context("when GetAllComponents returns an error", func() {
		It("should return nil BOM", func() {
			fakeK8sClient.GetAllComponentsStub = func(ctx context.Context) ([]model.Component, []string, error) {
				return nil, []string{}, assert.AnError
			}

			bom, err := GenerateBOM(fakeK8sClient)

			Expect(bom).To(BeNil())
			Expect(err).ToNot(BeNil())
		})
	})
})

var _ = Describe("Logger", Label("unit"), func() {
	Context("when give a log message", func() {
		It("should log", func() {
			log.Info().Msg("Hello World")
			log.Info().Msgf("Hello World from %s", "Person A")
			log.Info().Str("Name1", "Person A").Str("Name2", "Person B").Msg("Bananas!")
			// Wrap the error so it can be marshaled as JSON
			testError := errors.New("my Error Message")
			type ErrorWrapper struct {
				Message string `json:"message"`
			}
			errPayload := ErrorWrapper{
				Message: testError.Error(),
			}
			// Any is for objects implementing json.Marshaller - normally you would use .Err
			log.Info().Any("error", errPayload).Msg("Any Message")
			// Err is the way to go. You can force an Error at the Info level
			log.Info().Err(testError).Msg("Info Err Message")
			// But normally have it at the Error Level.
			log.Err(testError).Msg("Err Message with no level")
		})
	})
})

var _ = Describe("InitializeFilterStruct", Label("unit"), func() {

	DescribeTable("when initialized with various inputs",
		func(jsonInput string, expected *model.Filter) {
			// Load a new filter instance for each test case
			inputFilter := loadFilterFromJSON(jsonInput)

			InitializeFilterStruct(inputFilter)

			Expect(inputFilter).To(Equal(expected))
		},
		// Test cases
		Entry("should set default values when filter is empty",
			`{
			}`,
			&model.Filter{NamespacedInclusions: []model.NamespacedInclusion{model.NamespacedInclusion{
				Namespaces: []string{"*"},
				Resources:  []string{"*"},
			}}, NonNamespacedInclusions: model.NonNamespacedInclusions{}},
		),
		Entry("should set default namespace if no namespaces provided",
			`{
				"namespaced-inclusions": [
					{
						"namespaces": []
					}
				]
			}`,
			&model.Filter{
				NamespacedInclusions: []model.NamespacedInclusion{
					{Namespaces: []string{"*"}, Resources: []string{"*"}},
				},
			},
		),
		Entry("should convert resources to lowercase when namespace is '*'",
			`{
				"namespaced-inclusions": [
					{
						"namespaces": ["*"],
						"resources": ["pod", "deployment"]
					}
				]
			}`,
			&model.Filter{
				NamespacedInclusions: []model.NamespacedInclusion{
					{Namespaces: []string{"*"}, Resources: []string{"pod", "deployment"}},
				},
			},
		),
		Entry("set default namespace to if any of the namespaces in namespacedInclusion is *",
			`{
				"non-namespaced-inclusions": 
					{
					"resources": ["Namespace"]
					},
				"namespaced-inclusions": [
					{
						"namespaces": ["test-ns", "*"],
						"resources": ["pod", "deployment"]
					}
				]
			}`,
			&model.Filter{NonNamespacedInclusions: model.NonNamespacedInclusions{
				Resources: []string{"Namespace"},
			},
				NamespacedInclusions: []model.NamespacedInclusion{
					{
						Namespaces: []string{"*"},
						Resources:  []string{"pod", "deployment"},
					},
				}},
		),
		Entry("should set default resources if no resources are provided and any of the namespaces in namespacedInclusion is *",
			`{
				"non-namespaced-inclusions": 
					{
						"resources": ["Namespace"]
					},
				"namespaced-inclusions": [
					{
						"namespaces": ["test-ns", "*"]
					}
				]
			}`,
			&model.Filter{NonNamespacedInclusions: model.NonNamespacedInclusions{
				Resources: []string{"Namespace"},
			},
				NamespacedInclusions: []model.NamespacedInclusion{
					{
						Namespaces: []string{"*"},
						Resources:  []string{"*"},
					},
				}},
		),
		Entry("should not reset other namespacedInclusion if any of the namespaces in any namespacedInclusion is `*`",
			`{
				"non-namespaced-inclusions": 
					{
						"resources": ["Namespace"]
					},
				"namespaced-inclusions": [
					{
						"namespaces": ["test-ns", "*"],
						"resources": ["pod", "deployment"]
					},
					{
						"namespaces": ["default"],
						"resources": ["service"]
					}
				]
			}`,
			&model.Filter{NonNamespacedInclusions: model.NonNamespacedInclusions{
				Resources: []string{"Namespace"},
			},
				NamespacedInclusions: []model.NamespacedInclusion{
					{
						Namespaces: []string{"*"},
						Resources:  []string{"pod", "deployment"},
					},
					{
						Namespaces: []string{"default"},
						Resources:  []string{"service"},
					},
				}},
		),
		Entry("should handle case when namespace/resource is nil",
			`{
				"namespaced-inclusions": [
					{
						"namespaces": null
					},
					{
						"namespaces": ["namespace1"],
						"resources": ["pod", "service"]
					}
				]
			}`,
			&model.Filter{
				NamespacedInclusions: []model.NamespacedInclusion{
					{Namespaces: []string{"*"}, Resources: []string{"*"}},
					{Namespaces: []string{"namespace1"}, Resources: []string{"pod", "service"}},
				},
			},
		),
	)
})

func loadFilterFromJSON(jsonData string) *model.Filter {
	var filter model.Filter
	err := json.Unmarshal([]byte(jsonData), &filter)
	Expect(err).NotTo(HaveOccurred()) // Ensure JSON is valid
	return &filter
}
