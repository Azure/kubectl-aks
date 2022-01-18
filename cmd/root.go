// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"os"

	"github.com/Azure/kubectl-az/cmd/utils"
	"github.com/spf13/cobra"
)

// Common flags for all subcommands
var commonFlags utils.CommonFlags

var rootCmd = &cobra.Command{
	Use:   "kubectl-az",
	Short: "Microsoft Azure CLI kubectl plugin",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
