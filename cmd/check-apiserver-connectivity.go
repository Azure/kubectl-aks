// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Azure/kubectl-aks/cmd/utils"
)

// connCheckCmd is kept as a deprecated alias for backward compatibility.
// The canonical path is now: kubectl-aks check verify apiserver-connectivity
var connCheckCmd = &cobra.Command{
	Use:          "check-apiserver-connectivity",
	Short:        "Check connectivity between the nodes and the Kubernetes API Server",
	Long:         "Deprecated: use 'kubectl-aks check verify apiserver-connectivity' instead.",
	Deprecated:   "use 'kubectl-aks check verify apiserver-connectivity' instead",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Hint: use 'kubectl-aks check verify apiserver-connectivity' instead.")
		// Delegate to the new check command, passing our context-bearing cmd.
		newCmd, _, err := checkCmd.Find([]string{"verify", "apiserver-connectivity"})
		if err != nil {
			return fmt.Errorf("finding new check command: %w", err)
		}
		// Pass cmd (the caller) so that cmd.Context() is available to the RunE.
		return newCmd.RunE(cmd, args)
	},
}

func init() {
	utils.AddNodeFlags(connCheckCmd)
	utils.AddCommonFlags(connCheckCmd, &commonFlags)
	rootCmd.AddCommand(connCheckCmd)
}
