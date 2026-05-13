// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/kubectl-aks/cmd/utils"
	"github.com/Azure/kubectl-aks/cmd/utils/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var showConfigCmd = &cobra.Command{
	Use:          "show",
	Short:        "Show the configuration",
	Args:         cobra.NoArgs,
	RunE:         showConfigCmdRun,
	SilenceUsage: true,
}

var useNodeCmd = &cobra.Command{
	Use:          "use-node",
	Short:        "Set the current node in the configuration",
	RunE:         useNodeCmdRun,
	SilenceUsage: true,
}

var useClusterCmd = &cobra.Command{
	Use:          "use-cluster",
	Short:        "Set the current cluster in the configuration",
	RunE:         useClusterCmdRun,
	SilenceUsage: true,
}

var unsetCurrentNodeCmd = &cobra.Command{
	Use:          "unset-current-node",
	Short:        "Unset the current node in the configuration",
	Args:         cobra.NoArgs,
	RunE:         unsetCurrentNodeCmdRun,
	SilenceUsage: true,
}

var unsetNodeCmd = &cobra.Command{
	Use:          "unset-node",
	Short:        "Unset a given node in the configuration",
	RunE:         unsetNodeCmdRun,
	SilenceUsage: true,
}

var unsetClusterCmd = &cobra.Command{
	Use:          "unset-cluster",
	Short:        "Remove a cluster and all its nodes from the configuration",
	RunE:         unsetClusterCmdRun,
	SilenceUsage: true,
}

var unsetAllCmd = &cobra.Command{
	Use:          "unset-all",
	Short:        "Unset all nodes in the configuration",
	Args:         cobra.NoArgs,
	RunE:         unsetAllCmdRun,
	SilenceUsage: true,
}

var setNodeCmd = &cobra.Command{
	Use:          "set-node",
	Short:        "Set a given node in the configuration",
	Long:         "Set a given node in the configuration. Also, node, resource id and VMSS instance information are mutually exclusive",
	RunE:         setNodeCmdRun,
	SilenceUsage: true,
}

var listClustersCmd = &cobra.Command{
	Use:          "list-clusters",
	Short:        "List all clusters in the configuration",
	Args:         cobra.NoArgs,
	RunE:         listClustersCmdRun,
	SilenceUsage: true,
}

var importCmd = importCmdCommand()

func init() {
	utils.AddCommonFlags(configCmd, &commonFlags)

	if commonFlags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(showConfigCmd, useNodeCmd, useClusterCmd, unsetCurrentNodeCmd, unsetNodeCmd, unsetClusterCmd, unsetAllCmd, setNodeCmd, listClustersCmd, importCmd)
	utils.AddNodeFlagsOnly(setNodeCmd)
}

func showConfigCmdRun(cmd *cobra.Command, args []string) error {
	return config.New().ShowConfig()
}

func useNodeCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s <node name>", cmd.CommandPath())
	}
	return config.New().UseNodeConfig(args[0])
}

func useClusterCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s <cluster name>", cmd.CommandPath())
	}
	return config.New().UseClusterConfig(args[0])
}

func unsetCurrentNodeCmdRun(cmd *cobra.Command, args []string) error {
	return config.New().UnsetCurrentNodeConfig()
}

func unsetNodeCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s <node name>", cmd.CommandPath())
	}
	return config.New().UnsetNodeConfig(args[0])
}

func unsetClusterCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s <cluster name>", cmd.CommandPath())
	}
	return config.New().UnsetClusterConfig(args[0])
}

func unsetAllCmdRun(cmd *cobra.Command, args []string) error {
	return config.New().UnsetAllConfig()
}

func listClustersCmdRun(cmd *cobra.Command, args []string) error {
	cfg := config.New()
	clusters, err := cfg.ListClusters()
	if err != nil {
		return err
	}
	currentCluster := cfg.CurrentClusterName()
	for _, c := range clusters {
		prefix := "  "
		if c == currentCluster {
			prefix = "* "
		}
		fmt.Fprintf(os.Stdout, "%s%s\n", prefix, c)
	}
	return nil
}

func setNodeCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s <node name>", cmd.CommandPath())
	}

	cfg := config.New()
	if nf := cmd.Flag(utils.NodeKey).Value.String(); nf != "" {
		return cfg.SetNodeConfigWithNodeFlag(args[0], nf)
	} else if rid := cmd.Flag(utils.ResourceIDKey).Value.String(); rid != "" {
		return cfg.SetNodeConfigWithResourceIDFlag(args[0], rid)
	} else {
		subID := cmd.Flag(utils.SubscriptionIDKey).Value.String()
		nrg := cmd.Flag(utils.NodeResourceGroupKey).Value.String()
		vmss := cmd.Flag(utils.VMSSKey).Value.String()
		insID := cmd.Flag(utils.VMSSInstanceIDKey).Value.String()
		return cfg.SetNodeConfigWithVMSSInfoFlag(args[0], subID, nrg, vmss, insID)
	}
}

func importCmdCommand() *cobra.Command {
	var subscriptionID string
	var resourceGroup string
	var clusterName string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import Kubernetes nodes in the configuration",
		Long: fmt.Sprintf("Import Kubernetes nodes in the configuration."+"\n\n"+
			"By default (--runtime azure-api), nodes are imported via the Azure API and require"+"\n"+
			"'--%s', '--%s' and '--%s' flags."+"\n\n"+
			"Use '--runtime kube-api' to import via the Kubernetes API (kubeconfig). The cluster"+"\n"+
			"name is detected from the current kubeconfig context.",
			utils.SubscriptionIDKey, utils.ResourceGroupKey, utils.ClusterNameKey),
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolveRuntimeFromConfig()

			var vms map[string]*utils.VirtualMachineScaleSetVM

			utils.DefaultSpinner.Start()
			utils.DefaultSpinner.Suffix = " Importing..."

			switch runtimeFlag {
			case RuntimeKubeAPI:
				v, err := utils.VirtualMachineScaleSetVMsViaKubeconfig()
				utils.DefaultSpinner.Stop()
				if err != nil {
					return fmt.Errorf("getting VMSS VMs via Kubernetes API: %w", err)
				}
				vms = v

				// Always use the kubeconfig context as the cluster name to
				// avoid importing nodes from one context under a different
				// cluster name.
				detectedName := utils.DetectKubeconfigClusterName()
				if clusterName != "" && detectedName != "" && clusterName != detectedName {
					logrus.Warnf("--cluster-name %q ignored; using kubeconfig context %q (use --runtime azure-api to import a specific cluster)", clusterName, detectedName)
				}
				if detectedName != "" {
					clusterName = detectedName
				}

			default: // azure-api
				if subscriptionID == "" || resourceGroup == "" || clusterName == "" {
					utils.DefaultSpinner.Stop()
					return fmt.Errorf("'--%s', '--%s' and '--%s' flags are required for azure-api runtime.\n"+
						"Alternatively, use '--runtime kube-api' to import via kubeconfig",
						utils.SubscriptionIDKey, utils.ResourceGroupKey, utils.ClusterNameKey)
				}
				v, err := utils.VirtualMachineScaleSetVMsViaAzureAPI(subscriptionID, resourceGroup, clusterName)
				utils.DefaultSpinner.Stop()
				if err != nil {
					return fmt.Errorf("getting VMSS VMs via Azure API: %w", err)
				}
				vms = v
			}

			cfg := config.New()

			if clusterName != "" {
				if err := cfg.SetClusterMetadata(clusterName, subscriptionID, resourceGroup); err != nil {
					return fmt.Errorf("setting cluster metadata for %s: %w", clusterName, err)
				}
				// Clear existing nodes so the import is a full sync.
				if err := cfg.ClearClusterNodes(clusterName); err != nil {
					return fmt.Errorf("clearing old nodes for %s: %w", clusterName, err)
				}
				for nn, vm := range vms {
					if err := cfg.SetClusterNodeConfigWithVMSSInfo(clusterName, nn, vm.SubscriptionID, vm.NodeResourceGroup, vm.VMScaleSet, vm.InstanceID); err != nil {
						return fmt.Errorf("setting node config for %s: %w", nn, err)
					}
				}
				clusters, _ := cfg.ListClusters()
				if len(clusters) == 1 {
					if err := cfg.UseClusterConfig(clusterName); err != nil {
						return fmt.Errorf("setting current cluster: %w", err)
					}
				}
				return nil
			}

			// Legacy fallback: no cluster name available (kube-api without detectable context).
			logrus.Warn("Could not detect cluster name from kubeconfig context; storing nodes in legacy format")
			for nn, vm := range vms {
				if err := cfg.SetNodeConfigWithVMSSInfoFlag(nn, vm.SubscriptionID, vm.NodeResourceGroup, vm.VMScaleSet, vm.InstanceID); err != nil {
					return fmt.Errorf("setting node config for %s: %w", nn, err)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&subscriptionID, utils.SubscriptionIDKey, "", "", "Subscription ID of the cluster")
	cmd.Flags().StringVarP(&resourceGroup, utils.ResourceGroupKey, "", "", "Resource group of the cluster")
	cmd.Flags().StringVarP(&clusterName, utils.ClusterNameKey, "", "", "Name of the cluster")

	return cmd
}
