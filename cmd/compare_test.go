package cmd_test

import (
	. "cluster-codex/cmd"
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
)

var _ = Describe("extractBOMToMap - Unit", Label("unit"), func() {
	Context("when extractBOMToMap is called", func() {
		var containerPurl = "pkg:oci/etcd@sha256:24bc64e911039ecf00e263be2161797c758b7d82403ca5516ab64047a477f737?namespace=default&repository_url=registry.k8s.io%2Fetcd&version=3.5.7-0"
		var applicationPurl = "pkg:k8s/FlowSchema/service-accounts?apiVersion=flowcontrol.apiserver.k8s.io%2Fv1beta3"
		var key = "default/registry.k8s.io/etcd"
		var data = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"components": []interface{}{
					map[string]interface{}{
						"type":    "container",
						"purl":    containerPurl,
						"name":    "registry.k8s.io/etcd",
						"version": "3.5.7-0",
						"properties": []interface{}{
							map[string]interface{}{"name": "clx:k8s:componentKind", "value": "Image"},
							map[string]interface{}{"name": "clx:k8s:componentNamespace", "value": "default"},
						},
					},
					map[string]interface{}{
						"type":    "application",
						"purl":    applicationPurl,
						"name":    "service-accounts",
						"version": "flowcontrol.apiserver.k8s.io/v1beta3",
						"properties": []interface{}{
							map[string]interface{}{"name": "clx:k8s:componentKind", "value": "FlowSchema"},
							map[string]interface{}{"name": "clx:k8s:componentNamespace", "value": "default"},
						},
					},
				},
			},
		}

		It("should extract components of the given type", func() {

			result := ExtractBOMToMap(data, "container")

			Expect(result).To(HaveLen(1))
			Expect(result).To(HaveKey("default/registry.k8s.io/etcd"))
			Expect(result[key]).To(HaveKeyWithValue(BOM_PROPERTY_CONTAINER_NAMESPACE, "default"))
			Expect(result[key]).To(HaveKeyWithValue(BOM_PROPERTY_VERSION, "3.5.7-0"))
			Expect(result[key]).To(HaveKeyWithValue(BOM_PROPERTY_NAME, "registry.k8s.io/etcd"))
			Expect(result[key]).To(HaveKeyWithValue("clx:k8s:componentKind", "Image"))
		})

		It("should extract application of the given type", func() {

			result := ExtractBOMToMap(data, "application")

			Expect(result).To(HaveLen(1))
			Expect(result).To(HaveKey(applicationPurl))
			Expect(result[applicationPurl]).To(HaveKeyWithValue(BOM_PROPERTY_CONTAINER_NAMESPACE, "default"))
			Expect(result[applicationPurl]).To(HaveKeyWithValue(BOM_PROPERTY_VERSION, "flowcontrol.apiserver.k8s.io/v1beta3"))
			Expect(result[applicationPurl]).To(HaveKeyWithValue(BOM_PROPERTY_NAME, "service-accounts"))
			Expect(result[applicationPurl]).To(HaveKeyWithValue("clx:k8s:componentKind", "FlowSchema"))
		})

		It("should return empty for invalid dataType", func() {
			result := ExtractBOMToMap(data, "invalid")
			Expect(result).To(BeEmpty())
		})

		It("should process multiple components of the same type", func() {
			data := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"components": []interface{}{
						map[string]interface{}{
							"type":    "container",
							"purl":    containerPurl,
							"name":    "registry.k8s.io/etcd",
							"version": "3.5.7-0",
							"properties": []interface{}{
								map[string]interface{}{"name": "clx:k8s:componentKind", "value": "Image"},
								map[string]interface{}{"name": "clx:k8s:componentNamespace", "value": "default"},
							},
						},
						map[string]interface{}{
							"type": "container",
							//"purl":    "pkg:k8s/ClusterRoleBinding/system:controller:certificate-controller?apiVersion=rbac.authorization.k8s.io%2Fv1",
							"purl":    "pkg:oci/docker-public/loftsh/jspolicy@sha256:12345678?namespace=default&ownerRef=Deployment%2Fjspolicy&repository_url=11111111.dkr.ecr.us-west-2.amazonaws.com%2Fdocker-public%2Floftsh%2Fjspolicy",
							"name":    "11111111.dkr.ecr.us-west-2.amazonaws.com/docker-public/loftsh/jspolicy",
							"version": "0.0.1",
							"properties": []interface{}{
								map[string]interface{}{"name": "clx:k8s:componentKind", "value": "Image"},
								map[string]interface{}{"name": "clx:k8s:componentNamespace", "value": "default"},
							},
						},
					},
				},
			}

			result := ExtractBOMToMap(data, "container")

			Expect(result).To(HaveLen(2))
			Expect(result).To(HaveKey("default/registry.k8s.io/etcd"))
			Expect(result).To(HaveKey("default/*/docker-public/loftsh/jspolicy"))
		})
	})
})

var _ = Describe("compareBOMData - Unit", Label("unit"), func() {
	Context("when compareBOMData is called", func() {
		var expected = map[string]map[string]string{
			"pkg:oci/library/nginx@sha256:def7ef7fb89393d88ba6632347672cbde03926256220c2e535e4585335b838a0?namespace=&repository_url=index.docker.io%2Flibrary%2Fnginx": {
				"clx:k8s:componentKind":      "Image",
				"clx:k8s:componentNamespace": "",
				"clx:k8s:component:name":     "index.docker.io/library/nginx",
				"clx:k8s:component:version":  "3.5.7-0",
				"clx:k8s:component:ownerRef": "Deployment/nginx",
			},
		}

		var actual = map[string]map[string]string{
			"pkg:oci/library/nginx@sha256:def7ef7fb89393d88ba6632347672cbde03926256220c2e535e4585335b838a0?namespace=&repository_url=index.docker.io%2Flibrary%2Fnginx": {
				"clx:k8s:componentKind":      "Image",
				"clx:k8s:componentNamespace": "",
				"clx:k8s:component:name":     "index.docker.io/library/nginx",
				"clx:k8s:component:version":  "3.5.7-1",
				"clx:k8s:component:ownerRef": "Deployment/external-secrets",
			},
		}

		It("should return a warning when an expected component is missing in actual", func() {
			actual := map[string]map[string]string{} // Empty actual
			warn, err := CompareBOMData(expected, actual)

			Expect(warn).To(HaveLen(1))
			Expect(warn[0].Name).To(Equal("/index.docker.io/library/nginx"))
			Expect(err).To(BeEmpty())
		})

		It("should return no warnings or errors when an actual component is missing in expected", func() {
			expected := map[string]map[string]string{} // Empty expected
			warn, err := CompareBOMData(expected, actual)

			Expect(warn).To(HaveLen(1))
			Expect(err).To(BeEmpty())
		})

		It("should return an error when properties do not match", func() {

			warn, err := CompareBOMData(expected, actual)

			Expect(warn).To(BeEmpty())
			Expect(err).To(HaveLen(1))
			Expect(err[0].Name).To(Equal("/index.docker.io/library/nginx"))
			Expect(err[0].Properties).To(ContainElement(ComparisonProperty{
				PropertyName: BOM_PROPERTY_VERSION,
				Expected:     "3.5.7-0",
				Actual:       "3.5.7-1",
			}))
			Expect(err[0].Properties).To(ContainElement(ComparisonProperty{
				PropertyName: BOM_PROPERTY_OWNERREF,
				Expected:     "Deployment/nginx",
				Actual:       "Deployment/external-secrets",
			}))
		})

		It("should return no warnings or errors when expected and actual match", func() {
			actual := map[string]map[string]string{
				"pkg:oci/library/nginx@sha256:def7ef7fb89393d88ba6632347672cbde03926256220c2e535e4585335b838a0?namespace=&repository_url=index.docker.io%2Flibrary%2Fnginx": {
					"clx:k8s:componentKind":      "Image",
					"clx:k8s:componentNamespace": "",
					"clx:k8s:component:name":     "index.docker.io/library/nginx",
					"clx:k8s:component:version":  "3.5.7-0",
					"clx:k8s:component:ownerRef": "Deployment/nginx",
				},
			}

			warn, err := CompareBOMData(expected, actual)

			Expect(warn).To(BeEmpty())
			Expect(err).To(BeEmpty())
		})
	})
})

var _ = Describe("compareBOM - Unit", Label("unit"), func() {
	Context("when there are no mismatches between the BOMs", func() {

		It("should print no errors found during the BOM comparison", func() {
			expected, _ := os.ReadFile("../test/compare/expected.json")

			actual, _ := os.ReadFile("../test/compare/actual-exact-match.json")

			var expectedStruct *unstructured.Unstructured // Expected BOM
			var actualStruct *unstructured.Unstructured   // Generated BOM on given cluster

			json.Unmarshal(expected, &expectedStruct)
			json.Unmarshal(actual, &actualStruct)
			CompareBOM(expectedStruct, actualStruct)
		})
	})
})
