package k8_test

import (
	"cluster-codex/cmd"
	"cluster-codex/internal/k8"
	"cluster-codex/internal/model"
	"cluster-codex/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakediscovery "k8s.io/client-go/discovery/fake"
	dynamicfakeclient "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"log"
)

// ✅ Defines all commonly used GVRs
var gvrs = map[string]schema.GroupVersionResource{
	"pods":              {Group: "", Version: "v1", Resource: "pods"},
	"services":          {Group: "", Version: "v1", Resource: "services"},
	"deployments":       {Group: "apps", Version: "v1", Resource: "deployments"},
	"namespaces":        {Group: "", Version: "v1", Resource: "namespaces"},
	"persistentvolumes": {Group: "", Version: "v1", Resource: "persistentvolumes"},
}

var kinds = map[string]string{
	"pods":              "Pod",
	"services":          "Service",
	"deployments":       "Deployment",
	"namespaces":        "Namespace",
	"persistentvolumes": "PersistentVolume",
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
				"kind":       kinds[resourceType], // Get corresponding kind for a resource name (e.g., pods -> Pod, deployments -> Deployment)
				"metadata": map[string]interface{}{
					"name": name,
				},
			},
		}
		if namespace != "" {
			obj.Object["metadata"].(map[string]interface{})["namespace"] = namespace
		}

		// ✅ Only add statuses for Pods
		if resourceType == "pods" {
			obj.Object["status"] = map[string]interface{}{
				"containerStatuses": []interface{}{
					map[string]interface{}{
						"name":  "main-container",
						"ready": true,
						"state": map[string]interface{}{
							"running": map[string]interface{}{},
						},
					},
				},
				"initContainerStatuses": []interface{}{
					map[string]interface{}{
						"name":  "init-container",
						"ready": true,
						"state": map[string]interface{}{
							"terminated": map[string]interface{}{
								"exitCode": int64(0),
							},
						},
					},
				},
				"ephemeralContainerStatuses": []interface{}{
					map[string]interface{}{
						"name":  "debug-container",
						"ready": false,
						"state": map[string]interface{}{
							"waiting": map[string]interface{}{
								"reason": "ImagePullBackOff",
							},
						},
					},
				},
			}
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

var _ = Describe("Kubernetes - Unit", Label("unit"), func() {
	mockNamespaceList := []string{"default", "kube-system"}
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
				gvrs["pods"]:              "PodList",
				gvrs["services"]:          "ServiceList",
				gvrs["deployments"]:       "DeploymentList",
				gvrs["namespaces"]:        "NamespaceList",
				gvrs["persistentvolumes"]: "PersistentvolumesList",
			},
		)

		// ✅ Create mock resources
		mockPods := createMockResources("pods", []string{"pod-1", "pod-2"}, "default")
		mockPods2 := createMockResources("pods", []string{"pod-3", "pod-4"}, "kube-system")
		mockDeployments := createMockResources("deployments", []string{"deployment-1"}, "default")
		mockNamespaces := createMockResources("namespaces", mockNamespaceList, "")
		mockPersistentVolumes := createMockResources("persistentvolumes", []string{"pv-1"}, "")

		// ✅ Add mock objects to FakeDynamicClient tracker
		for _, pod := range mockPods {
			_, err := fakeDynamicClient.Resource(gvrs["pods"]).Namespace("default").Create(context.TODO(), &pod, v1.CreateOptions{})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
		for _, pod := range mockPods2 {
			_, err := fakeDynamicClient.Resource(gvrs["pods"]).Namespace("kube-system").Create(context.TODO(), &pod, v1.CreateOptions{})
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
		for _, pv := range mockPersistentVolumes {
			_, err := fakeDynamicClient.Resource(gvrs["persistentvolumes"]).Create(context.TODO(), &pv, v1.CreateOptions{})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}

		// ✅ Convert unstructured Pods to corev1.Pod and add to fake client with Status
		mockPods = append(mockPods, mockPods2...)
		for _, unstructuredPod := range mockPods {
			var pod corev1.Pod
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredPod.Object, &pod)
			if err != nil {
				log.Fatalf("Error converting unstructured to pod: %v", err)
			}

			// ✅ Set Pod Spec with Containers
			pod.Spec.Containers = []corev1.Container{
				{
					Name:  "main-container",
					Image: "nginx:latest",
				},
			}
			pod.Spec.InitContainers = []corev1.Container{
				{
					Name:  "init-container",
					Image: "busybox:latest",
				},
			}
			pod.Spec.EphemeralContainers = []corev1.EphemeralContainer{
				{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{
						Name:  "debug-container",
						Image: "busybox:debug",
					},
				},
			}

			// ✅ Set ContainerStatuses
			pod.Status.ContainerStatuses = []corev1.ContainerStatus{
				{
					Name:  "main-container",
					Ready: true,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					},
				},
			}

			// ✅ Set InitContainerStatuses
			pod.Status.InitContainerStatuses = []corev1.ContainerStatus{
				{
					Name:  "init-container",
					Ready: true,
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
			}

			// ✅ Set EphemeralContainerStatuses
			pod.Status.EphemeralContainerStatuses = []corev1.ContainerStatus{
				{
					Name:  "debug-container",
					Ready: false,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "ImagePullBackOff",
						},
					},
				},
			}

			// ✅ First, add the pod (spec only)
			err = fakeClientset.Tracker().Add(&pod)
			if err != nil {
				log.Fatalf("Error adding pod to fake client: %v", err)
			}

			// ✅ Then, update the pod status using Tracker().Update() — simulating a real status update
			err = fakeClientset.Tracker().Update(corev1.SchemeGroupVersion.WithResource("pods"), &pod, pod.Namespace)
			if err != nil {
				log.Fatalf("Error updating pod status in fake client: %v", err)
			}
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
						{Name: "persistentvolumes", Namespaced: false, Kind: "PersistentVolume"},
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
		DescribeTable("should return all the components in the cluster",
			func(namespaces []string, expectedComponents int, includeKubeSystem bool, nonNamespacedResources []string) {
				// Setup the filter with the namespaces
				k8.K8Filter = model.Filter{
					NamespacedInclusions:    []model.NamespacedInclusion{{Namespaces: namespaces}},
					NonNamespacedInclusions: model.NonNamespacedInclusions{Resources: nonNamespacedResources},
				}

				cmd.InitializeFilterStruct(&k8.K8Filter)

				components, namespaces, err := fakeK8sClient.GetAllComponents(context.Background())

				Expect(err).To(BeNil())
				// ✅ Assert correct number of components (Pods + Deployments + Namespaces + PersistentVolumes)
				Expect(len(components)).To(Equal(expectedComponents)) // pod-1, pod-2, pod-3, pod-4, deployment-1, default, kube-system, pv-1

				// ✅ Convert list into a map for easy lookup
				componentMap := make(map[string]model.Component)
				for _, comp := range components {
					componentMap[comp.Name] = comp
				}

				// ✅ Assert specific components exist with correct types
				Expect(componentMap).To(HaveKey("pod-1"))
				Expect(componentMap).To(HaveKey("pod-2"))
				if includeKubeSystem {
					Expect(componentMap).To(HaveKey("pod-3"))
					Expect(componentMap).To(HaveKey("pod-4"))
				}
				Expect(componentMap).To(HaveKey("deployment-1"))

				// Assert only if Namespace Kind is included in the NonNamespacedInclusion
				if len(nonNamespacedResources) == 0 || utils.Contains(nonNamespacedResources, "Namespace") {
					Expect(componentMap).To(HaveKey("default"))     // Namespace
					Expect(componentMap).To(HaveKey("kube-system")) // Namespace
					// ✅ Check for namespace handling
					Expect(componentMap["default"].Type).To(Equal("application"))
					if includeKubeSystem {
						Expect(componentMap["kube-system"].Type).To(Equal("application"))
					}
					// ✅ Check properties for Namespaces (They should NOT have a model.ComponentNamespace property)
					Expect(componentMap["default"].Properties).To(ContainElement(
						model.Property{Name: model.ComponentKind, Values: []string{"Namespace"}},
					))
					Expect(componentMap["default"].Properties).ToNot(ContainElement(
						model.Property{Name: model.ComponentNamespace, Values: []string{"default"}}, // Namespaces shouldn't have this property
					))

					Expect(componentMap["kube-system"].Properties).To(ContainElement(
						model.Property{Name: model.ComponentKind, Values: []string{"Namespace"}},
					))
					Expect(componentMap["kube-system"].Properties).ToNot(ContainElement(
						model.Property{Name: model.ComponentNamespace, Values: []string{"kube-system"}},
					))
				}

				// ✅ Assert individual component details
				Expect(componentMap["pod-1"].Type).To(Equal("application"))
				Expect(componentMap["pod-1"].Name).To(Equal("pod-1"))
				Expect(componentMap["pod-1"].Version).To(Equal("v1")) // No version for pods
				Expect(componentMap["pod-1"].PackageURL).To(Equal(fmt.Sprintf("%s:%s/Pod/pod-1?apiVersion=v1&namespace=default", model.PkgPrefix, model.K8sPrefix)))

				Expect(componentMap["deployment-1"].Type).To(Equal("application"))
				Expect(componentMap["deployment-1"].Name).To(Equal("deployment-1"))
				Expect(componentMap["deployment-1"].PackageURL).To(Equal(fmt.Sprintf("%s:%s/Deployment/deployment-1?apiVersion=apps%%2Fv1&namespace=default", model.PkgPrefix, model.K8sPrefix)))

				//Assert non-namespaced PersistentVolume
				Expect(componentMap["pv-1"].Type).To(Equal("application"))
				Expect(componentMap["pv-1"].Name).To(Equal("pv-1"))
				Expect(componentMap["pv-1"].PackageURL).To(Equal(fmt.Sprintf("%s:%s/PersistentVolume/pv-1?apiVersion=v1", model.PkgPrefix, model.K8sPrefix)))

				// ✅ Check properties for Pods (Namespace & Kind)
				Expect(componentMap["pod-1"].Properties).To(ContainElements(
					model.Property{Name: model.ComponentKind, Values: []string{"Pod"}},
					model.Property{Name: model.ComponentNamespace, Values: []string{"default"}},
				))
				Expect(componentMap["pod-2"].Properties).To(ContainElements(
					model.Property{Name: model.ComponentKind, Values: []string{"Pod"}},
					model.Property{Name: model.ComponentNamespace, Values: []string{"default"}},
				))
				if includeKubeSystem {
					Expect(componentMap["pod-3"].Properties).To(ContainElements(
						model.Property{Name: model.ComponentKind, Values: []string{"Pod"}},
						model.Property{Name: model.ComponentNamespace, Values: []string{"kube-system"}},
					))
					Expect(componentMap["pod-4"].Properties).To(ContainElements(
						model.Property{Name: model.ComponentKind, Values: []string{"Pod"}},
						model.Property{Name: model.ComponentNamespace, Values: []string{"kube-system"}},
					))
				}
				// ✅ Check properties for Deployments (Namespace & Kind)
				Expect(componentMap["deployment-1"].Properties).To(ContainElements(
					model.Property{Name: model.ComponentKind, Values: []string{"Deployment"}},
					model.Property{Name: model.ComponentNamespace, Values: []string{"default"}},
				))

				// ✅ Check properties for non namespaced kind PersistentVolume (They Should NOT have ComponentNamespace property)
				Expect(componentMap["pv-1"].Properties).To(ContainElement(
					model.Property{Name: model.ComponentKind, Values: []string{"PersistentVolume"}},
				))
				Expect(componentMap["pv-1"].Properties).ToNot(ContainElement(
					model.Property{Name: model.ComponentNamespace, Values: []string{"default"}}, // Non-namespaced resources shouldn't have this property
				))

				// ✅ Ensure no licenses/hashes exist since they were not set in mock data
				Expect(componentMap["pod-1"].Licenses).To(BeEmpty())
				Expect(componentMap["pod-1"].Hashes).To(BeEmpty())

				// Make sure that there are no services components - we didn't give any mock data for them.
				for _, comp := range components {
					kind, found := comp.GetProperty(model.ComponentKind)
					Expect(found && kind == "Services").ToNot(BeTrue(), "Expected not to find a 'Service', but found one")
				}
			},
			Entry("No Namespaces", []string{}, 8, true, []string{}),
			Entry("Specific Test Namespaces", mockNamespaceList, 8, true, []string{}),
			Entry("Specific Test Namespace", []string{"default"}, 6, false, []string{}),
			Entry("Specific Test Namespace and Non-namespaced resources", []string{"default"}, 6, false, []string{"Namespace", "PersistentVolume"}),
			Entry("Specific Test Namespace and Non-namespaced resources (not including `Namespace`)", []string{"default"}, 4, false, []string{"PersistentVolume"}),
			Entry("Specific Test Namespace and Non-namespaced resources (All)", []string{"default"}, 6, false, []string{"*"}),
		)
	})

	Context("when GetAllImages is called with a K8s client", func() {
		It("should return all the images in the cluster", func() {

			components, err := fakeK8sClient.GetAllImages(context.Background(), mockNamespaceList)

			Expect(err).To(BeNil())
			// ✅ Assert correct number of components (Pods + Deployments + Namespaces)
			Expect(len(components)).To(Equal(3)) // busybox:latest, busybox:debug,nginx:latest

			// ✅ Convert list into a map for easy lookup
			componentMap := make(map[string]model.Component)
			for _, comp := range components {
				componentMap[comp.Name+":"+comp.Version] = comp
			}

			// ✅ Assert specific components exist with correct types
			Expect(componentMap).To(HaveKey("index.docker.io/library/busybox:latest"))
			Expect(componentMap).To(HaveKey("index.docker.io/library/busybox:debug"))
			Expect(componentMap).To(HaveKey("index.docker.io/library/nginx:latest"))

			// ✅ Assert individual component details
			Expect(componentMap["index.docker.io/library/busybox:latest"].PackageURL).To(Equal("pkg:oci/library/busybox?repository_url=index.docker.io%2Flibrary%2Fbusybox&version=latest"))
			Expect(componentMap["index.docker.io/library/busybox:debug"].PackageURL).To(Equal("pkg:oci/library/busybox?repository_url=index.docker.io%2Flibrary%2Fbusybox&version=debug"))
			Expect(componentMap["index.docker.io/library/nginx:latest"].PackageURL).To(Equal("pkg:oci/library/nginx?repository_url=index.docker.io%2Flibrary%2Fnginx&version=latest"))
			componentPointer := componentMap["index.docker.io/library/nginx:latest"]
			property, found := componentPointer.GetPropertyObject(model.ComponentNamespace)
			Expect(found).To(Equal(true))
			Expect(property).ToNot(BeNil())
			Expect(len(property.Values)).To(Equal(2))
			Expect(property.Values).To(ConsistOf(mockNamespaceList))
		})
	})

})

var _ = Describe("GetAppPkgId", Label("unit"), func() {
	It("should generate the correct URL when namespace is provided", func() {
		result := k8.GetAppPkgId("deployment", "my-app", "default", "apps/v1")
		expected := fmt.Sprintf("%s:%s/deployment/my-app?apiVersion=apps%%2Fv1&namespace=default", model.PkgPrefix, model.K8sPrefix)
		Expect(result).To(Equal(expected))
	})

	It("should generate the correct URL when namespace is empty", func() {
		result := k8.GetAppPkgId("service", "my-service", "", "v1")
		expected := fmt.Sprintf("%s:%s/service/my-service?apiVersion=v1", model.PkgPrefix, model.K8sPrefix)
		Expect(result).To(Equal(expected))
	})

	It("should correctly encode special characters in apiVersion", func() {
		result := k8.GetAppPkgId("APIService", "v1beta1.metrics.k8s.io", "test", "apiregistration.k8s.io/v1")
		expected := fmt.Sprintf("%s:%s/APIService/v1beta1.metrics.k8s.io?apiVersion=apiregistration.k8s.io%%2Fv1&namespace=test", model.PkgPrefix, model.K8sPrefix)
		Expect(result).To(Equal(expected))
	})

	It("should correctly encode special characters in namespace", func() {
		result := k8.GetAppPkgId("configmap", "my-config", "kube-system", "v1")
		expected := fmt.Sprintf("%s:%s/configmap/my-config?apiVersion=v1&namespace=kube-system", model.PkgPrefix, model.K8sPrefix)
		Expect(result).To(Equal(expected))
	})
})

func loadFilterFromJSON(jsonData string) *model.Filter {
	var filter model.Filter
	err := json.Unmarshal([]byte(jsonData), &filter)
	Expect(err).NotTo(HaveOccurred()) // Ensure JSON is valid
	return &filter
}
