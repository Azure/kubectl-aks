// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"

	"github.com/Azure/kubectl-aks/cmd/utils"
	"github.com/spf13/cobra"
)

var (
	command string
	timeout int
)

var runCommandCmd = &cobra.Command{
	Use:          "run-command",
	Short:        "Run a command in a node",
	RunE:         runCommandCmdRun,
	SilenceUsage: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("usage: %s <command>", cmd.CommandPath())
		}
		command = args[0]

		return nil
	},
}

func init() {
	runCommandCmd.Flags().IntVar(&timeout, "timeout", utils.DefaultRunCommandTimeoutInSeconds, "timeout in seconds for the command to complete")
	utils.AddNodeFlags(runCommandCmd)
	utils.AddCommonFlags(runCommandCmd, &commonFlags)
	rootCmd.AddCommand(runCommandCmd)
}

func runCommandCmdRun(cmd *cobra.Command, args []string) error {
	cred, err := utils.GetCredentials()
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	vm, err := utils.VirtualMachineScaleSetVMFromConfig()
	if err != nil {
		return fmt.Errorf("getting vm: %w", err)
	}

	res, err := utils.RunCommand(cmd.Context(), cred, vm, &command, commonFlags.Verbose, &timeout)
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	fmt.Printf("\n%s", res)
	return nil
}
