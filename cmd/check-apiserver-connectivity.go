// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Azure/kubectl-aks/cmd/utils"
	"github.com/spf13/cobra"
)

var connCheckCmd = &cobra.Command{
	Use:          "check-apiserver-connectivity",
	Short:        "Check connectivity between the nodes and the Kubernetes API Server",
	RunE:         connCheckCmdRun,
	SilenceUsage: true,
}

func init() {
	utils.AddNodeFlags(connCheckCmd)
	utils.AddCommonFlags(connCheckCmd, &commonFlags)
	rootCmd.AddCommand(connCheckCmd)
}

func connCheckCmdRun(cmd *cobra.Command, args []string) error {
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
	command := "kubectl --kubeconfig /var/lib/kubelet/kubeconfig version > /dev/null; echo $?"
	res, err := utils.RunCommand(cmd.Context(), cred, vm, &command, commonFlags.Verbose, nil, utils.OutputTruncateTail)
	if err != nil {
		return fmt.Errorf("failed to run command that checks connectivity: %w", err)
	}

	// Extract stdout and stderr from response.
	// Expected format: "[stdout]<text>[stderr]<text>"
	split := regexp.MustCompile(`(\[(stdout|stderr)\])`).Split(res, -1)
	if len(split) != 3 {
		return fmt.Errorf("couldn't parse response message:\n%s", res)
	}
	stdOutput := strings.TrimSpace(split[1])
	stdError := strings.TrimSpace(split[2])

	// The stdout should contain the returned value of "kubectl version":
	// 0 (succeeded), otherwise (failure)
	ret, err := strconv.Atoi(stdOutput)
	if err != nil {
		return fmt.Errorf("couldn't parse stdout of response message:\n%s", res)
	}
	if ret != 0 {
		fmt.Printf("\nConnectivity check: failed with returned value %d: %s\n",
			ret, stdError)

		// Force the binary to return an exit code != 0 (forwarding command's
		// return value). Useful if it is used in scripts.
		os.Exit(ret)
	}

	fmt.Println("\nConnectivity check: succeeded")

	return nil
}
