package k8_test

import (
	"cluster-codex/internal/k8"
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

var _ = Describe("Kubernetes", Label("unittest"), func() {
	var (
		fakeK8sClient     *k8.K8sClient
		fakeClientset     *fake.Clientset
		fakeDynamicClient *dynamicfakeclient.FakeDynamicClient
	)

	BeforeEach(func() {
		// Initialize fake Kubernetes clientset
		fakeClientset = fake.NewSimpleClientset()

		// Initialize fake dynamic client with custom list kinds
		fakeDynamicClient = dynamicfakeclient.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{
				// Define the custom GVR for the resources you're testing
				{Group: "", Version: "v1", Resource: "pods"}:            "PodList",
				{Group: "apps", Version: "v1", Resource: "deployments"}: "DeploymentList",
			},
		)

		// Initialize K8sClient with fake clients
		fakeK8sClient = &k8.K8sClient{
			K8sContext:    "test-cluster",
			Config:        nil,
			Client:        fakeClientset,     // Fake Clientset
			DynamicClient: fakeDynamicClient, // Fake Dynamic Client
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
