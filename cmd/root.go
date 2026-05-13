// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Azure/kubectl-aks/cmd/utils"
	"github.com/Azure/kubectl-aks/cmd/utils/config"
)

const (
	RuntimeAzureAPI = "azure-api"
	RuntimeKubeAPI  = "kube-api"

	runtimeKey    = "runtime"
	debugImageKey = "debug-image"
)

// Common flags for all subcommands
var commonFlags utils.CommonFlags

// Runtime flags
var (
	runtimeFlag string
	debugImage  string
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-aks",
	Short: "Azure Kubernetes Service (AKS) kubectl plugin",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		if cfg.IsLegacyConfig() {
			fmt.Fprintln(os.Stderr, "⚠ Your config uses the old format (top-level 'nodes' without 'clusters').")
			fmt.Fprintln(os.Stderr, "  Please run 'kubectl-aks config unset-all' and re-import with 'kubectl-aks config import'.")
			fmt.Fprintln(os.Stderr)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&runtimeFlag, runtimeKey, RuntimeAzureAPI,
		"Runtime to use for command execution. Supported values: azure-api, kube-api")
	rootCmd.PersistentFlags().StringVar(&debugImage, debugImageKey, "busybox:latest",
		"Container image to use for the kube-api runtime")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
