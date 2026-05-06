// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/Azure/kubectl-aks/cmd/utils"
	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
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
	rt, err := buildRuntime()
	if err != nil {
		return err
	}

	// Check connectivity by executing "kubectl version" on the node.
	command := "KUBECTL=$(command -v kubectl || echo kubectl); [ -x /opt/bin/kubectl ] && KUBECTL=/opt/bin/kubectl; $KUBECTL --kubeconfig /var/lib/kubelet/kubeconfig version > /dev/null; echo -n $?"
	opts := &pkgruntime.RunOptions{
		NodeName: utils.GetNodeName(),
		Command:  command,
		Timeout:  utils.DefaultRunCommandTimeoutInSeconds,
	}

	res, err := rt.RunCommand(cmd.Context(), opts)
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
		fmt.Printf("Connectivity check: failed with returned value %d: %s\n",
			ret, res.Stderr)

		// Force the binary to return an exit code != 0 (forwarding command's
		// return value). Useful if it is used in scripts.
		os.Exit(ret)
	}

	fmt.Println("Connectivity check: succeeded")

	return nil
}
