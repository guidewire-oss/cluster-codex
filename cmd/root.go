package cmd

import (
	"cluster-codex/internal/k8"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/version"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "clx",
	Short: "clx - Kubernetes Bill of Materials",
}

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate KBOM for the provided K8s cluster",
	RunE:  runGenerate,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	fmt.Println("This is root command")
	rootCmd.AddCommand(GenerateCmd)
}

func runGenerate(cmd *cobra.Command, _ []string) error {
	k8sClient, err := k8.GetClient()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}
	var serverVersion *version.Info
	serverVersion, err = k8sClient.Client.Discovery().ServerVersion()
	if err != nil {
		log.Fatalf("Failed to get server version: %v", err)
	}

	fmt.Printf("Git version:%s", serverVersion.String())
	return err
}
