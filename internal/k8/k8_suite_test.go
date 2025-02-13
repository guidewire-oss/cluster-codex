package k8_test

import (
	"cluster-codex/internal/k8"
	"context"
	"github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakediscovery "k8s.io/client-go/discovery/fake"
	dynamicfakeclient "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestK8(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8 Suite")
}

// Custom FakeDiscovery that overrides ServerPreferredResources
type CustomFakeDiscovery struct {
	fakediscovery.FakeDiscovery
	Resources []*v1.APIResourceList
}

// Override ServerPreferredResources to return our mock resources
func (c *CustomFakeDiscovery) ServerPreferredResources() ([]*v1.APIResourceList, error) {
	return c.Resources, nil
}

var _ = Describe("Kubernetes", Label("unittest"), func() {
	var (
		fakeK8sClient     *k8.K8sClient
		fakeClientset     *fake.Clientset
		fakeDynamicClient *dynamicfakeclient.FakeDynamicClient
		fakeDiscovery     *CustomFakeDiscovery
		podGVR            schema.GroupVersionResource
		mockPods          []unstructured.Unstructured
	)

	BeforeEach(func() {
		// Initialize fake Kubernetes clientset
		fakeClientset = fake.NewSimpleClientset()

		podGVR = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

		// Initialize fake dynamic client with custom list kinds
		fakeDynamicClient = dynamicfakeclient.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{
				// Define the custom GVR for the resources you're testing
				{Group: "", Version: "v1", Resource: "pods"}: "PodList",
				//{Group: "", Version: "v1", Resource: "services"}:         "ServiceList",
				//{Group: "", Version: "apps/v1", Resource: "deployments"}: "DeploymentList",
				//{Group: "apps", Version: "v1", Resource: "deployments"}:  "DeploymentList",
			},
		)

		// Define mock pod resources
		mockPods = []unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod-1",
						"namespace": "default",
					},
				},
			},
			{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "pod-2",
						"namespace": "default",
					},
				},
			},
		}

		// Add mock pods to the fake client tracker
		for _, pod := range mockPods {
			_, err := fakeDynamicClient.Resource(podGVR).Namespace("default").Create(context.TODO(), &pod, v1.CreateOptions{})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}

		fakeDiscovery = &CustomFakeDiscovery{
			FakeDiscovery: fakediscovery.FakeDiscovery{
				Fake: &fakeClientset.Fake,
			},
		}

		// Replace ServerPreferredResources to return mock resources
		resources := []*v1.APIResourceList{
			{
				GroupVersion: "v1",
				APIResources: []v1.APIResource{
					{Name: "pods", Namespaced: true, Kind: "Pod"},
					//{Name: "services", Namespaced: true, Kind: "Service"},
				},
			},
			//{
			//	GroupVersion: "apps/v1",
			//	APIResources: []v1.APIResource{
			//		{Name: "deployments", Namespaced: true, Kind: "Deployment"},
			//	},
			//},
		}
		fakeDiscovery.Resources = resources

		// Initialize K8sClient with fake clients
		fakeK8sClient = &k8.K8sClient{
			K8sContext:    "test-cluster",
			Config:        nil,
			Client:        fakeClientset,     // Fake Clientset
			DynamicClient: fakeDynamicClient, // Fake Dynamic Client
			Discovery:     fakeDiscovery,
		}
	})

	Context("when GetAllComponents is called with a K8s client ", func() {
		It("should return all the components in cluster", func() {

			components, err := fakeK8sClient.GetAllComponents(context.Background())

			Expect(err).To(BeNil())
			Expect(components).To(HaveLen(0))

		})
	})
})
