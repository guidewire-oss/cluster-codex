package k8

import (
	"context"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
	"testing"
	"time"
)

func TestAccessToKubernetes(t *testing.T) {
	client, err := GetClient()

	namespace := "test-namespace"

	// Ensure namespace exists
	_, err = client.Client.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: namespace},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Logf("Namespace %s might already exist: %v", namespace, err)
	}
	// Cleanup namespace
	defer func() {
		t.Logf("Cleaning up namespace %s", namespace)
		_ = client.Client.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{})
	}()

	// Load the Deployment YAML
	deployment, err := loadYAML("deployment-test.yaml")
	if err != nil {
		t.Fatalf("Failed to load YAML file: %v", err)
	}

	// Apply the Deployment
	_, err = client.Client.AppsV1().Deployments(deployment.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create deployment: %v", err)
	}
	t.Log("Deployment created successfully")

	// Wait for the Deployment to be available
	timeout := time.After(30 * time.Second)
	ticker := time.Tick(2 * time.Second)

	for {
		select {
		case <-timeout:
			t.Fatalf("Deployment did not become available in time")
		case <-ticker:
			deploy, err := client.Client.AppsV1().Deployments(deployment.Namespace).Get(context.TODO(), deployment.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("Failed to get deployment: %v", err)
			}
			if deploy.Status.AvailableReplicas > 0 {
				t.Log("Deployment is available")
				return
			}
		}
	}

}

// loadYAML loads and parses a YAML file
func loadYAML(filePath string) (*appsv1.Deployment, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var deployment appsv1.Deployment
	err = yaml.Unmarshal(data, &deployment)
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}
