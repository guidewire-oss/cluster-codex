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

//func TestGenerateBom(t *testing.T) {
//	k8sClient, err := k8.GetClient()
//	if err != nil {
//		log.Fatalf("Error creating Kubernetes client: %v", err)
//	}
//	bom := generateBOM(k8sClient)
//	assert.NotEmpty(t, bom)
//	assert.Positive(t, len(bom.Components))
//	foundNamespace := false
//	foundPod := false
//	for _, component := range bom.Components {
//		kind, found := component.GetProperty("clx:k8s:componentKind")
//		if found && kind == "Namespace" && component.Name == "test" {
//			foundNamespace = true
//		}
//		if found && kind == "Pod" && strings.HasPrefix(component.Name, "nginx-deployment") {
//			foundPod = true
//		}
//	}
//	assert.True(t, foundNamespace, "Expected to find namespace in bom but did not")
//	assert.True(t, foundPod, "Expected to find pod in bom but did not")
//}

//func TestListResources(t *testing.T) {
//	//t.Skip("Skip, used for debugging locally.")
//	ctx := context.Background()
//	k8sClient, err := k8.GetClient()
//	if err != nil {
//		log.Fatalf("Error creating Kubernetes client: %v", err)
//	}
//
//	gvr := schema.GroupVersionResource{
//		Group:    "",
//		Version:  "v1",
//		Resource: "endpoints",
//	}
//
//	k8sResources, k8serr := k8sClient.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
//	if k8serr != nil {
//		log.Fatalf("Failed to list resources %v %v", gvr, k8serr)
//	}
//
//	if k8sResources == nil || len(k8sResources.Items) == 0 {
//		log.Fatalf("No resources found for GVR %v", gvr)
//	}
//
//	for _, item := range k8sResources.Items {
//		log.Println("########################################")
//		log.Printf("Name: %s\n", item.GetName())
//		log.Printf("Kind: %s\n", item.GetKind())
//		log.Printf("Namespace: %s\n", item.GetNamespace())
//		log.Printf("Labels: %s\n", item.GetLabels())
//		log.Println("########################################")
//	}
//
//	log.Printf("Printed %d resources\n", len(k8sResources.Items))
//
//	for _, item := range k8sResources.Items {
//		log.Printf("Name: %s\n", item.GetName())
//	}
//}

//func TestGenerateBOMSuccess(t *testing.T) {
//	// Create the fake K8sClientInterface using Counterfeiter
//	fakeK8sClient := new(k8fakes.FakeK8sClientInterface)
//
//	// Mock the GetAllComponents method
//	fakeK8sClient.GetAllComponentsStub = func(ctx context.Context) ([]model.Component, error) {
//
//		mockComponent := model.Component{
//			Type:       "application",
//			Name:       "test-component",
//			Version:    "1.0.0",
//			PackageURL: "pkg:docker/test-component@1.0.0", // Optional (omit if not needed)
//			Properties: []model.Property{
//				{Name: "clx:k8s:componentKind", Value: "HelmChart"},
//				{Name: "clx:k8s:namespace", Value: "flux-system"},
//				{Name: "clx:k8s:componentVersion", Value: "1.0.0"},
//			},
//		}
//		mockResponse := []model.Component{mockComponent}
//		return mockResponse, nil
//	}
//
//	// Call GenerateBOM with the fake client
//	bom := GenerateBOM(fakeK8sClient)
//
//	// Assert that the BOM is not nil and contains the expected components
//	assert.NotNil(t, bom)
//	assert.Len(t, bom.Components, 1)
//	assert.Equal(t, "test-component", bom.Components[0].Name)
//}
//
//func TestGenerateBOMError(t *testing.T) {
//	// Create the fake K8sClientInterface using Counterfeiter
//	fakeK8sClient := new(k8fakes.FakeK8sClientInterface)
//
//	// Mock the GetAllComponents method
//	fakeK8sClient.GetAllComponentsStub = func(ctx context.Context) ([]model.Component, error) {
//
//		return nil, assert.AnError
//	}
//
//	bom := GenerateBOM(fakeK8sClient)
//	assert.Nil(t, bom)
//}

var _ = Describe("GenerateBOM", Label("unittest"), func() {
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
					{Name: "clx:k8s:componentKind", Value: "HelmChart"},
					{Name: "clx:k8s:namespace", Value: "flux-system"},
					{Name: "clx:k8s:componentVersion", Value: "1.0.0"},
				},
			}
			mockResponse := []model.Component{mockComponent}
			fakeK8sClient.GetAllComponentsStub = func(ctx context.Context) ([]model.Component, error) {
				return mockResponse, nil
			}

			bom := GenerateBOM(fakeK8sClient)

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

			bom := GenerateBOM(fakeK8sClient)

			Expect(bom).To(BeNil())
		})
	})
})
