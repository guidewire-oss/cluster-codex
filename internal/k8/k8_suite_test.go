package k8_test

import (
	"cluster-codex/internal/k8"
	"cluster-codex/internal/model"
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

// ✅ Defines all commonly used GVRs
var gvrs = map[string]schema.GroupVersionResource{
	"pods":        {Group: "", Version: "v1", Resource: "pods"},
	"services":    {Group: "", Version: "v1", Resource: "services"},
	"deployments": {Group: "apps", Version: "v1", Resource: "deployments"},
	"namespaces":  {Group: "", Version: "v1", Resource: "namespaces"},
}

// ✅ Generates mock Kubernetes resources dynamically
func createMockResources(resourceType string, names []string, namespace string) []unstructured.Unstructured {
	var resources []unstructured.Unstructured
	for _, name := range names {
		var apiVersion string
		if gvrs[resourceType].Group == "" {
			apiVersion = gvrs[resourceType].Version
		} else {
			apiVersion = gvrs[resourceType].Group + "/" + gvrs[resourceType].Version
		}
		obj := unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": apiVersion,
				"kind":       capitalize(resourceType), // Auto capitalize kind (e.g., Pod, Deployment)
				"metadata": map[string]interface{}{
					"name": name,
				},
			},
		}
		if namespace != "" {
			obj.Object["metadata"].(map[string]interface{})["namespace"] = namespace
		}
		resources = append(resources, obj)
	}
	return resources
}

// ✅ Custom FakeDiscovery that overrides ServerPreferredResources
type CustomFakeDiscovery struct {
	fakediscovery.FakeDiscovery
	Resources []*v1.APIResourceList
}

// ✅ Override ServerPreferredResources to return mock API resources
func (c *CustomFakeDiscovery) ServerPreferredResources() ([]*v1.APIResourceList, error) {
	return c.Resources, nil
}

var _ = Describe("Kubernetes", Label("unit"), func() {
	var (
		fakeK8sClient     *k8.K8sClient
		fakeClientset     *fake.Clientset
		fakeDynamicClient *dynamicfakeclient.FakeDynamicClient
		fakeDiscovery     *CustomFakeDiscovery
	)

	BeforeEach(func() {
		// ✅ Initialize fake Kubernetes clientset
		fakeClientset = fake.NewSimpleClientset()

		// ✅ Initialize fake dynamic client with auto-detected resources
		fakeDynamicClient = dynamicfakeclient.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{
				gvrs["pods"]:        "PodList",
				gvrs["services"]:    "ServiceList",
				gvrs["deployments"]: "DeploymentList",
				gvrs["namespaces"]:  "NamespaceList",
			},
		)

		// ✅ Create mock resources
		mockPods := createMockResources("pods", []string{"pod-1", "pod-2"}, "default")
		mockDeployments := createMockResources("deployments", []string{"deployment-1"}, "default")
		mockNamespaces := createMockResources("namespaces", []string{"default", "kube-system"}, "")

		// ✅ Add mock objects to FakeDynamicClient tracker
		for _, pod := range mockPods {
			_, err := fakeDynamicClient.Resource(gvrs["pods"]).Namespace("default").Create(context.TODO(), &pod, v1.CreateOptions{})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
		for _, deployment := range mockDeployments {
			_, err := fakeDynamicClient.Resource(gvrs["deployments"]).Namespace("default").Create(context.TODO(), &deployment, v1.CreateOptions{})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
		for _, namespace := range mockNamespaces {
			_, err := fakeDynamicClient.Resource(gvrs["namespaces"]).Create(context.TODO(), &namespace, v1.CreateOptions{})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}

		// ✅ Initialize FakeDiscovery
		fakeDiscovery = &CustomFakeDiscovery{
			FakeDiscovery: fakediscovery.FakeDiscovery{
				Fake: &fakeClientset.Fake,
			},
			Resources: []*v1.APIResourceList{
				{
					GroupVersion: "v1",
					APIResources: []v1.APIResource{
						{Name: "pods", Namespaced: true, Kind: "Pod"},
						{Name: "services", Namespaced: true, Kind: "Service"},
						{Name: "namespaces", Namespaced: false, Kind: "Namespace"},
					},
				},
				{
					GroupVersion: "apps/v1",
					APIResources: []v1.APIResource{
						{Name: "deployments", Namespaced: true, Kind: "Deployment"},
					},
				},
			},
		}

		// ✅ Initialize K8sClient with fake clients
		fakeK8sClient = &k8.K8sClient{
			K8sContext:    "test-cluster",
			Config:        nil,
			Client:        fakeClientset,     // Fake Clientset
			DynamicClient: fakeDynamicClient, // Fake Dynamic Client
			Discovery:     fakeDiscovery,
		}
	})

	Context("when GetAllComponents is called with a K8s client", func() {
		It("should return all the components in the cluster", func() {
			components, err := fakeK8sClient.GetAllComponents(context.Background())

			Expect(err).To(BeNil())
			// ✅ Assert correct number of components (Pods + Deployments + Namespaces)
			Expect(len(components)).To(Equal(5)) // pod-1, pod-2, deployment-1, default, kube-system

			// ✅ Convert list into a map for easy lookup
			componentMap := make(map[string]model.Component)
			for _, comp := range components {
				componentMap[comp.Name] = comp
			}

			// ✅ Assert specific components exist with correct types
			Expect(componentMap).To(HaveKey("pod-1"))
			Expect(componentMap).To(HaveKey("pod-2"))
			Expect(componentMap).To(HaveKey("deployment-1"))
			Expect(componentMap).To(HaveKey("default"))     // Namespace
			Expect(componentMap).To(HaveKey("kube-system")) // Namespace

			// ✅ Assert individual component details
			Expect(componentMap["pod-1"].Type).To(Equal("application"))
			Expect(componentMap["pod-1"].Name).To(Equal("pod-1"))
			Expect(componentMap["pod-1"].Version).To(Equal("v1")) // No version for pods
			Expect(componentMap["pod-1"].PackageURL).To(BeEmpty())

			Expect(componentMap["deployment-1"].Type).To(Equal("application"))
			Expect(componentMap["deployment-1"].Name).To(Equal("deployment-1"))

			// ✅ Check for namespace handling
			Expect(componentMap["default"].Type).To(Equal("application"))
			Expect(componentMap["kube-system"].Type).To(Equal("application"))

			// ✅ Check properties for Pods (Namespace & Kind)
			Expect(componentMap["pod-1"].Properties).To(ContainElements(
				model.Property{Name: "clx:k8s:componentKind", Value: "Pods"},
				model.Property{Name: "clx:k8s:namespace", Value: "default"},
			))
			Expect(componentMap["pod-2"].Properties).To(ContainElements(
				model.Property{Name: "clx:k8s:componentKind", Value: "Pods"},
				model.Property{Name: "clx:k8s:namespace", Value: "default"},
			))

			// ✅ Check properties for Deployments (Namespace & Kind)
			Expect(componentMap["deployment-1"].Properties).To(ContainElements(
				model.Property{Name: "clx:k8s:componentKind", Value: "Deployments"},
				model.Property{Name: "clx:k8s:namespace", Value: "default"},
			))

			// ✅ Check properties for Namespaces (They should NOT have a "clx:k8s:namespace" property)
			Expect(componentMap["default"].Properties).To(ContainElement(
				model.Property{Name: "clx:k8s:componentKind", Value: "Namespaces"},
			))
			Expect(componentMap["default"].Properties).ToNot(ContainElement(
				model.Property{Name: "clx:k8s:namespace", Value: "default"}, // Namespaces shouldn't have this property
			))

			Expect(componentMap["kube-system"].Properties).To(ContainElement(
				model.Property{Name: "clx:k8s:componentKind", Value: "Namespaces"},
			))
			Expect(componentMap["kube-system"].Properties).ToNot(ContainElement(
				model.Property{Name: "clx:k8s:namespace", Value: "kube-system"},
			))

			// ✅ Ensure no licenses/hashes exist since they were not set in mock data
			Expect(componentMap["pod-1"].Licenses).To(BeEmpty())
			Expect(componentMap["pod-1"].Hashes).To(BeEmpty())

			// Make sure that there are no services components - we didn't give any mock data for them.
			for _, comp := range components {
				kind, found := comp.GetProperty("clx:k8s:componentKind")
				Expect(found && kind == "Services").ToNot(BeTrue(), "Expected not to find a 'Service', but found one")
			}
		})
	})
})

// ✅ Helper function to capitalize resource type (e.g., "pods" → "Pod")
func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return string(s[0]-32) + s[1:]
}
