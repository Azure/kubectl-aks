// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"
	"os"

	"github.com/Azure/kubectl-aks/cmd/utils"
	"github.com/spf13/cobra"
)

var (
	command      string
	timeout      int
	truncateHead bool
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

	// We want to truncate the tail by default because most of commands used for
	// debugging print a column header which is necessary to understand the
	// output. In addition, if the output is too long, those tools usually
	// provide mechanisms for filtering. Notice it is the opposite behaviour of
	// the Azure CLI.
	runCommandCmd.PersistentFlags().BoolVarP(
		&truncateHead,
		"truncate-head", "",
		false,
		"the output will be always truncated at the tail to return the first 4096 bytes by default, "+
			"this flag allows to return the latest 4096 bytes instead",
	)

	rootCmd.AddCommand(runCommandCmd)
}

func runCommandCmdRun(cmd *cobra.Command, args []string) error {
	cred, err := utils.GetCredentials()
	if err != nil {
		return fmt.Errorf("authenticating: %w", err)
	}

	vm, err := utils.VirtualMachineScaleSetVMFromConfig()
	if err != nil {
		return fmt.Errorf("getting vm: %w", err)
	}

	outputTruncate := utils.OutputTruncateTail
	if truncateHead {
		outputTruncate = utils.OutputTruncateHead
	}

	res, err := utils.RunCommand(cmd.Context(), cred, vm, &command, commonFlags.Verbose, &timeout, outputTruncate)
	if err != nil {
		return fmt.Errorf("running command: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s", res.Stderr)
	fmt.Fprintf(os.Stdout, "%s", res.Stdout)
	return nil
}
