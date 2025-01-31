package k8

import (
	"fmt"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

type k8sClient struct {
	K8sContext    string
	Config        *rest.Config
	Client        kubernetes.Interface
	DynamicClient dynamic.Interface
}

// NewClientset takes a path to a kubeconfig file and returns a Kubernetes clientset.
func GetClient() (*k8sClient, error) {

	kubeConfigPath := os.Getenv("KUBECONFIG")
	fmt.Printf("Reading kubeconfig from %s\n", kubeConfigPath)

	if kubeConfigPath == "" {
		kubeConfigPath = os.Getenv("HOME") + "/.kube/config"
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Printf("Error creating config: %v", err)
		return nil, err
	}

	// Create the clientset from the config.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Error creating clientset: %v", err)
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Printf("Error creating dynamic client: %v", err)
		return nil, err
	}

	return &k8sClient{
		K8sContext:    "default",
		Config:        config,
		Client:        clientset,
		DynamicClient: dynamicClient,
	}, nil
}
