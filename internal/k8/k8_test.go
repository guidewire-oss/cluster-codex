package k8

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"testing"
)

func TestAccessToKubernetes(t *testing.T) {
	client, err := GetClient()
	if err != nil {
		t.Skip(fmt.Printf("Error running GetClient: %v, skipping this test.", err))
	}

	namespace := "atmos-system"
	// List pods in the namespace

	// Define the GVR (Group, Version, Resource) for Pods
	gvr := schema.GroupVersionResource{
		Group:    "", // Core API group
		Version:  "v1",
		Resource: "pods",
	}
	podList, err := client.DynamicClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing pods: %v", err)
	}

	// Print pod names
	fmt.Printf("Pods in %s namespace:\n", namespace)
	for _, pod := range podList.Items {
		fmt.Printf("- %s\n", pod.GetName())
	}
}
