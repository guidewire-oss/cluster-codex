package cmd

import (
	"cluster-codex/internal/k8"
	"context"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"strings"
	"testing"
)

func TestGenerateBom(t *testing.T) {
	k8sClient, err := k8.GetClient()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}
	bom := generateBOM(k8sClient)
	assert.NotEmpty(t, bom)
	assert.Positive(t, len(bom.Components))
	foundNamespace := false
	foundPod := false
	for _, component := range bom.Components {
		kind, found := component.GetProperty("clx:k8s:componentKind")
		if found && kind == "Namespace" && component.Name == "test" {
			foundNamespace = true
		}
		if found && kind == "Pod" && strings.HasPrefix(component.Name, "nginx-deployment") {
			foundPod = true
		}
	}
	assert.True(t, foundNamespace, "Expected to find namespace in bom but did not")
	assert.True(t, foundPod, "Expected to find pod in bom but did not")
}

func TestListResources(t *testing.T) {
	t.Skip("Skip, used for debugging locally.")
	ctx := context.Background()
	k8sClient, err := k8.GetClient()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "endpoints",
	}

	k8sResources, k8serr := k8sClient.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if k8serr != nil {
		log.Fatalf("Failed to list resources %v %v", gvr, k8serr)
	}

	if k8sResources == nil || len(k8sResources.Items) == 0 {
		log.Fatalf("No resources found for GVR %v", gvr)
	}

	for _, item := range k8sResources.Items {
		log.Println("########################################")
		log.Printf("Name: %s\n", item.GetName())
		log.Printf("Kind: %s\n", item.GetKind())
		log.Printf("Namespace: %s\n", item.GetNamespace())
		log.Printf("Labels: %s\n", item.GetLabels())
		log.Println("########################################")
	}

	log.Printf("Printed %d resources\n", len(k8sResources.Items))

	for _, item := range k8sResources.Items {
		log.Printf("Name: %s\n", item.GetName())
	}
}
