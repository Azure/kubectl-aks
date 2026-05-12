// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/Azure/kubectl-aks/cmd/utils"
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
