// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Azure/kubectl-aks/cmd/utils"
	"github.com/Azure/kubectl-aks/cmd/utils/config"
	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
	"github.com/Azure/kubectl-aks/pkg/runtime/vmss"
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
	// Fan-out: check connectivity across all nodes in a cluster
	if cl := utils.GetClusterFlag(); cl != "" {
		return connCheckOnCluster(cmd, cl)
	}

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

func connCheckOnCluster(cmd *cobra.Command, clusterName string) error {
	cfg := config.New()
	nodes, err := cfg.ListClusterNodes(clusterName)
	if err != nil {
		return fmt.Errorf("listing nodes for cluster %s: %w", clusterName, err)
	}
	if len(nodes) == 0 {
		return fmt.Errorf("cluster %q has no nodes", clusterName)
	}

	type checkResult struct {
		NodeName string
		Success  bool
		Err      error
	}

	results := make([]checkResult, len(nodes))
	var wg sync.WaitGroup

	connCommand := "KUBECTL=$(command -v kubectl || echo kubectl); [ -x /opt/bin/kubectl ] && KUBECTL=/opt/bin/kubectl; $KUBECTL --kubeconfig /var/lib/kubelet/kubeconfig version > /dev/null; echo -n $?"

	for i, nodeName := range nodes {
		wg.Add(1)
		go func(idx int, nn string) {
			defer wg.Done()
			cr := checkResult{NodeName: nn}

			nc, ok := cfg.GetClusterNodeConfig(clusterName, nn)
			if !ok {
				cr.Err = fmt.Errorf("node %q not found in cluster %q", nn, clusterName)
				results[idx] = cr
				return
			}

			resolveRuntimeFromConfig()

			var rt pkgruntime.Runtime
			switch runtimeFlag {
			case RuntimeKubeAPI:
				r, buildErr := buildKubectlDebugRuntime()
				if buildErr != nil {
					cr.Err = buildErr
					results[idx] = cr
					return
				}
				rt = r
			default:
				cred, credErr := utils.GetCredentials()
				if credErr != nil {
					cr.Err = fmt.Errorf("authenticating: %w", credErr)
					results[idx] = cr
					return
				}

				vm := &utils.VirtualMachineScaleSetVM{
					SubscriptionID:    nc.GetString(utils.SubscriptionIDKey),
					NodeResourceGroup: nc.GetString(utils.NodeResourceGroupKey),
					VMScaleSet:        nc.GetString(utils.VMSSKey),
					InstanceID:        nc.GetString(utils.VMSSInstanceIDKey),
				}

				rt = &vmss.Runtime{
					Credential:     cred,
					VM:             vm,
					OutputTruncate: utils.OutputTruncateTail,
				}
			}

			opts := &pkgruntime.RunOptions{
				NodeName: nn,
				Command:  connCommand,
				Timeout:  utils.DefaultRunCommandTimeoutInSeconds,
			}

			res, runErr := rt.RunCommand(cmd.Context(), opts)
			if runErr != nil {
				cr.Err = runErr
				results[idx] = cr
				return
			}

			ret, parseErr := strconv.Atoi(res.Stdout)
			if parseErr != nil {
				cr.Err = fmt.Errorf("couldn't parse stdout: %s", res.Stdout)
				results[idx] = cr
				return
			}
			cr.Success = ret == 0
			results[idx] = cr
		}(i, nodeName)
	}
	wg.Wait()

	hasFailure := false
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(os.Stdout, "%-40s ERROR: %s\n", r.NodeName, r.Err)
			hasFailure = true
		} else if r.Success {
			fmt.Fprintf(os.Stdout, "%-40s Connectivity check: succeeded\n", r.NodeName)
		} else {
			fmt.Fprintf(os.Stdout, "%-40s Connectivity check: failed\n", r.NodeName)
			hasFailure = true
		}
	}

	if hasFailure {
		os.Exit(1)
	}
	return nil
}
