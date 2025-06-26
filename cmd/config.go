// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"

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

var importCmd = importCmdCommand()

func init() {
	utils.AddCommonFlags(configCmd, &commonFlags)

	if commonFlags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(showConfigCmd, useNodeCmd, unsetCurrentNodeCmd, unsetNodeCmd, unsetAllCmd, setNodeCmd, importCmd)
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

func unsetCurrentNodeCmdRun(cmd *cobra.Command, args []string) error {
	return config.New().UnsetCurrentNodeConfig()
}

func unsetNodeCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s <node name>", cmd.CommandPath())
	}
	return config.New().UnsetNodeConfig(args[0])
}

func unsetAllCmdRun(cmd *cobra.Command, args []string) error {
	return config.New().UnsetAllConfig()
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

	virtualMachineScaleSetVMs := func() (map[string]*utils.VirtualMachineScaleSetVM, error) {
		utils.DefaultSpinner.Start()
		utils.DefaultSpinner.Suffix = " Importing..."

		if subscriptionID != "" && resourceGroup != "" && clusterName != "" {
			vms, err := utils.VirtualMachineScaleSetVMsViaAzureAPI(subscriptionID, resourceGroup, clusterName)
			utils.DefaultSpinner.Stop()
			if err != nil {
				return nil, fmt.Errorf("getting VMSS VMs via Azure API: %w", err)
			}
			return vms, nil
		}

		vms, err := utils.VirtualMachineScaleSetVMsViaKubeconfig()
		utils.DefaultSpinner.Stop()
		if err != nil {
			logrus.Warn("Could not get VMSS VMs via Kubernetes API")
			logrus.Warnf("Please provide '--%s', '--%s' and '--%s' flags to get VMSS VMs via Azure API",
				utils.SubscriptionIDKey, utils.ResourceGroupKey, utils.ClusterNameKey)
			return nil, fmt.Errorf("getting VMSS VMs via Kuberntes API: %w", err)
		}

		return vms, nil
	}

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import Kubernetes nodes in the configuration",
		Long: fmt.Sprintf("Import Kubernetes nodes in the configuration"+"\n\n"+
			"It uses kubeconfig by default, but it can also use Azure API to get VMSS VMs."+"\n"+
			"In case of Azure API, you need to provide '--%s', '--%s' and '--%s' flags.",
			utils.SubscriptionIDKey, utils.ResourceGroupKey, utils.ClusterNameKey),
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			vms, err := virtualMachineScaleSetVMs()
			if err != nil {
				return err
			}
			cfg := config.New()
			for nn, vm := range vms {
				if err = cfg.SetNodeConfigWithVMSSInfoFlag(nn, vm.SubscriptionID, vm.NodeResourceGroup, vm.VMScaleSet, vm.InstanceID); err != nil {
					return fmt.Errorf("setting node config for %s: %w", nn, err)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&subscriptionID, utils.SubscriptionIDKey, "", "", "Subscription ID of the cluster (only needed with Azure API)")
	cmd.Flags().StringVarP(&resourceGroup, utils.ResourceGroupKey, "", "", "Resource group of the cluster (only needed with Azure API)")
	cmd.Flags().StringVarP(&clusterName, utils.ClusterNameKey, "", "", "Name of the cluster (only needed with Azure API)")

	return cmd
}
