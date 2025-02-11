package cmd

import (
	"cluster-codex/internal/k8"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"testing"
)

func TestListOneResource(t *testing.T) {
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

	namespace := "keda-system"

	k8sResources, k8serr := k8sClient.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if k8serr != nil {
		log.Fatalf("Failed to list resources in namespace %v %s %v", gvr, namespace, k8serr)
	}

	if k8sResources == nil || len(k8sResources.Items) == 0 {
		log.Fatalf("No resources found for GVR %v in namespace %s", gvr, namespace)
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
