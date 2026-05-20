// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Azure/kubectl-aks/cmd/utils"
	"github.com/Azure/kubectl-aks/cmd/utils/config"
	"github.com/Azure/kubectl-aks/pkg/check"
	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
	"github.com/Azure/kubectl-aks/pkg/runtime/vmss"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run diagnostic checks on AKS nodes",
	Long: `Run diagnostic checks on AKS nodes.

Available check modes:
  verify    Point-in-time checks that run once and report pass/fail
  trace     Duration-based checks that observe the node for a specified period

Examples:
  kubectl-aks check verify apiserver-connectivity
  kubectl-aks check verify dns-resolution --node mynode
  kubectl-aks check trace dns --duration 30 --node mynode
  kubectl-aks check trace tcp-drops --duration 60 --cluster mycluster`,
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Run point-in-time diagnostic checks",
}

var traceCmd = &cobra.Command{
	Use:   "trace",
	Short: "Run duration-based diagnostic checks that observe the node over time",
}

// traceDuration holds the --duration flag value for trace checks.
var traceDuration int

func init() {
	// Register the check parent command
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(verifyCmd)
	checkCmd.AddCommand(traceCmd)

	// Add --duration flag to trace command
	traceCmd.PersistentFlags().IntVar(&traceDuration, "duration", check.DefaultTraceDuration,
		"Duration in seconds to run the trace")

	// Auto-register all checks as subcommands
	for _, c := range check.All() {
		registerCheck(c)
	}
}

func registerCheck(c check.Check) {
	cmd := &cobra.Command{
		Use:          c.Name(),
		Short:        c.Description(),
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE:         makeCheckRunFunc(c),
	}

	utils.AddNodeFlags(cmd)
	utils.AddCommonFlags(cmd, &commonFlags)

	switch c.Mode() {
	case check.ModeVerify:
		verifyCmd.AddCommand(cmd)
	case check.ModeTrace:
		traceCmd.AddCommand(cmd)
	}
}

func makeCheckRunFunc(c check.Check) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		duration := check.DefaultTraceDuration
		if c.Mode() == check.ModeTrace {
			duration = traceDuration
		}

		// Cluster fan-out mode
		if cl := utils.GetClusterFlag(); cl != "" {
			return runCheckOnCluster(cmd, c, cl, duration)
		}

		// Single node mode
		rt, err := buildRuntime()
		if err != nil {
			return err
		}

		nr, err := check.RunOnNode(cmd.Context(), c, rt, utils.GetNodeName(), utils.DefaultRunCommandTimeoutInSeconds, duration)
		if err != nil {
			return err
		}

		output, hasFailure := check.FormatResults([]check.NodeResult{*nr})
		fmt.Fprint(os.Stdout, output)
		if hasFailure {
			os.Exit(1)
		}
		return nil
	}
}

func runCheckOnCluster(cmd *cobra.Command, c check.Check, clusterName string, duration int) error {
	cfg := config.New()
	nodes, err := cfg.ListClusterNodes(clusterName)
	if err != nil {
		return fmt.Errorf("listing nodes for cluster %s: %w", clusterName, err)
	}
	if len(nodes) == 0 {
		return fmt.Errorf("cluster %q has no nodes", clusterName)
	}

	factory := func(nn string) (pkgruntime.Runtime, error) {
		nc, ok := cfg.GetClusterNodeConfig(clusterName, nn)
		if !ok {
			return nil, fmt.Errorf("node %q not found in cluster %q", nn, clusterName)
		}

		resolveRuntimeFromConfig()

		switch runtimeFlag {
		case RuntimeKubeAPI:
			return buildKubectlDebugRuntime()
		default:
			cred, err := utils.GetCredentials()
			if err != nil {
				return nil, fmt.Errorf("authenticating: %w", err)
			}
			vm := &utils.VirtualMachineScaleSetVM{
				SubscriptionID:    nc.GetString(utils.SubscriptionIDKey),
				NodeResourceGroup: nc.GetString(utils.NodeResourceGroupKey),
				VMScaleSet:        nc.GetString(utils.VMSSKey),
				InstanceID:        nc.GetString(utils.VMSSInstanceIDKey),
			}
			return &vmss.Runtime{
				Credential:     cred,
				VM:             vm,
				OutputTruncate: utils.OutputTruncateTail,
			}, nil
		}
	}

	results := check.RunOnNodes(cmd.Context(), c, nodes, factory, utils.DefaultRunCommandTimeoutInSeconds, duration)
	output, hasFailure := check.FormatResults(results)
	fmt.Fprint(os.Stdout, output)
	if hasFailure {
		os.Exit(1)
	}
	return nil
}
