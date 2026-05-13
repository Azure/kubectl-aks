// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/Azure/kubectl-aks/cmd/utils"
	"github.com/Azure/kubectl-aks/cmd/utils/config"
	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
	"github.com/Azure/kubectl-aks/pkg/runtime/kubectldebug"
	"github.com/Azure/kubectl-aks/pkg/runtime/vmss"
	"github.com/kinvolk/inspektor-gadget/pkg/k8sutil"
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
	// Fan-out: run across all nodes in a cluster
	if cl := utils.GetClusterFlag(); cl != "" {
		return runCommandOnCluster(cmd, cl)
	}

	rt, err := buildRuntime()
	if err != nil {
		return err
	}

	opts := &pkgruntime.RunOptions{
		NodeName: utils.GetNodeName(),
		Command:  command,
		Timeout:  timeout,
	}

	res, err := rt.RunCommand(cmd.Context(), opts)
	if err != nil {
		return fmt.Errorf("running command: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s", res.Stderr)
	fmt.Fprintf(os.Stdout, "%s", res.Stdout)
	return nil
}

// nodeResult holds the output from a single node execution.
type nodeResult struct {
	NodeName string
	Stdout   string
	Stderr   string
	Err      error
}

func runCommandOnCluster(cmd *cobra.Command, clusterName string) error {
	cfg := config.New()
	nodes, err := cfg.ListClusterNodes(clusterName)
	if err != nil {
		return fmt.Errorf("listing nodes for cluster %s: %w", clusterName, err)
	}
	if len(nodes) == 0 {
		return fmt.Errorf("cluster %q has no nodes", clusterName)
	}

	results := make([]nodeResult, len(nodes))
	var wg sync.WaitGroup

	for i, nodeName := range nodes {
		wg.Add(1)
		go func(idx int, nn string) {
			defer wg.Done()
			results[idx] = runOnSingleNode(cmd, cfg, clusterName, nn)
		}(i, nodeName)
	}
	wg.Wait()

	for _, r := range results {
		fmt.Fprintf(os.Stdout, "=== %s ===\n", r.NodeName)
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", r.Err)
		} else {
			if r.Stderr != "" {
				fmt.Fprintf(os.Stderr, "%s", r.Stderr)
			}
			fmt.Fprintf(os.Stdout, "%s", r.Stdout)
		}
		fmt.Fprintln(os.Stdout)
	}
	return nil
}

func runOnSingleNode(cmd *cobra.Command, cfg *config.Config, clusterName, nodeName string) nodeResult {
	nr := nodeResult{NodeName: nodeName}

	nc, ok := cfg.GetClusterNodeConfig(clusterName, nodeName)
	if !ok {
		nr.Err = fmt.Errorf("node %q not found in cluster %q", nodeName, clusterName)
		return nr
	}

	resolveRuntimeFromConfig()

	var rt pkgruntime.Runtime

	switch runtimeFlag {
	case RuntimeKubeAPI:
		r, err := buildKubectlDebugRuntime()
		if err != nil {
			nr.Err = err
			return nr
		}
		rt = r
	default: // azure-api
		cred, err := utils.GetCredentials()
		if err != nil {
			nr.Err = fmt.Errorf("authenticating: %w", err)
			return nr
		}

		vm := &utils.VirtualMachineScaleSetVM{
			SubscriptionID:    nc.GetString(utils.SubscriptionIDKey),
			NodeResourceGroup: nc.GetString(utils.NodeResourceGroupKey),
			VMScaleSet:        nc.GetString(utils.VMSSKey),
			InstanceID:        nc.GetString(utils.VMSSInstanceIDKey),
		}

		outputTruncate := utils.OutputTruncateTail
		if truncateHead {
			outputTruncate = utils.OutputTruncateHead
		}

		rt = &vmss.Runtime{
			Credential:     cred,
			VM:             vm,
			OutputTruncate: outputTruncate,
		}
	}

	opts := &pkgruntime.RunOptions{
		NodeName: nodeName,
		Command:  command,
		Timeout:  timeout,
	}

	res, err := rt.RunCommand(cmd.Context(), opts)
	if err != nil {
		nr.Err = err
		return nr
	}
	nr.Stdout = res.Stdout
	nr.Stderr = res.Stderr
	return nr
}

// buildRuntime creates the appropriate runtime based on the --runtime flag.
// It checks (1) CLI flag, (2) config file for runtime preference.
func buildRuntime() (pkgruntime.Runtime, error) {
	resolveRuntimeFromConfig()

	switch runtimeFlag {
	case RuntimeAzureAPI:
		return buildVMSSRuntime()
	case RuntimeKubeAPI:
		return buildKubectlDebugRuntime()
	default:
		return nil, fmt.Errorf("unsupported runtime %q: use %q or %q",
			runtimeFlag, RuntimeAzureAPI, RuntimeKubeAPI)
	}
}

func buildVMSSRuntime() (pkgruntime.Runtime, error) {
	cred, err := utils.GetCredentials()
	if err != nil {
		return nil, fmt.Errorf("authenticating: %w", err)
	}

	vm, err := utils.VirtualMachineScaleSetVMFromConfig()
	if err != nil {
		return nil, fmt.Errorf("getting vm: %w", err)
	}

	outputTruncate := utils.OutputTruncateTail
	if truncateHead {
		outputTruncate = utils.OutputTruncateHead
	}

	return &vmss.Runtime{
		Credential:     cred,
		VM:             vm,
		OutputTruncate: outputTruncate,
	}, nil
}

func buildKubectlDebugRuntime() (pkgruntime.Runtime, error) {
	config, err := utils.KubernetesConfigFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("getting kubernetes config: %w", err)
	}

	clientset, err := k8sutil.NewClientsetFromConfigFlags(utils.KubernetesConfigFlags)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes client: %w", err)
	}

	return &kubectldebug.Runtime{
		Clientset: clientset,
		Config:    config,
		Image:     debugImage,
	}, nil
}
