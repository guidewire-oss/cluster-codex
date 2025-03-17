package cmd

import (
	"cluster-codex/internal/config"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var logLevel string
var rootCmd = &cobra.Command{
	Use:   "clx",
	Short: "clx - Kubernetes Bill of Materials",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		config.ConfigureLogger(logLevel)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(GenerateCmd)
	rootCmd.AddCommand(CompareCmd)
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "warn", "Set the logging level (debug, info, warn, error)")
}
