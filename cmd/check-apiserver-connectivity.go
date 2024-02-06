// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Azure/kubectl-aks/cmd/utils"
	"github.com/spf13/cobra"
)

var connCheckCmd = &cobra.Command{
	Use:          "check-apiserver-connectivity",
	Short:        "Check connectivity between the nodes and the Kubernetes API Server",
	Args:         cobra.NoArgs,
	RunE:         connCheckCmdRun,
	SilenceUsage: true,
}

func init() {
	utils.AddNodeFlags(connCheckCmd)
	utils.AddCommonFlags(connCheckCmd, &commonFlags)
	rootCmd.AddCommand(connCheckCmd)
}

func connCheckCmdRun(cmd *cobra.Command, args []string) error {
	utils.DefaultSpinner.Start()
	cred, err := utils.GetCredentials()
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	vm, err := utils.VirtualMachineScaleSetVMFromConfig()
	if err != nil {
		return fmt.Errorf("getting vm: %w", err)
	}

	// Check connectivity by executing "kubectl version" on the node. This
	// command will try to contact the API server to get the Kubernetes version
	// it is running. Use only the return value of the command, tough.
	command := "kubectl --kubeconfig /var/lib/kubelet/kubeconfig version > /dev/null; echo -n $?"
	res, err := utils.RunCommand(cmd.Context(), cred, vm, &command, commonFlags.Verbose, nil, utils.OutputTruncateTail)
	if err != nil {
		return fmt.Errorf("failed to run command that checks connectivity: %w", err)
	}

	// The stdout should contain the returned value of "kubectl version":
	// 0 (succeeded), otherwise (failure)
	ret, err := strconv.Atoi(res.Stdout)
	if err != nil {
		return fmt.Errorf("couldn't parse stdout of response message:\n%s", res.Stdout)
	}
	if ret != 0 {
		fmt.Printf("\nConnectivity check: failed with returned value %d: %s\n",
			ret, res.Stderr)

		// Force the binary to return an exit code != 0 (forwarding command's
		// return value). Useful if it is used in scripts.
		os.Exit(ret)
	}

	utils.DefaultSpinner.Stop()
	fmt.Println("Connectivity check: succeeded")

	return nil
}
