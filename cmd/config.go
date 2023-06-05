// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"

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

var importCmd = &cobra.Command{
	Use:          "import",
	Short:        "Import Kubernetes nodes in the configuration",
	RunE:         importCmdRun,
	SilenceUsage: true,
}

func init() {
	utils.AddCommonFlags(configCmd, &commonFlags)
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

func importCmdRun(cmd *cobra.Command, args []string) error {
	vms, err := utils.VirtualMachineScaleSetVMsViaKubeconfig()
	if err != nil {
		return fmt.Errorf("failed to get VMSS VMs: %w", err)
	}

	cfg := config.New()
	for nn, vm := range vms {
		if err = cfg.SetNodeConfigWithVMSSInfoFlag(nn, vm.SubscriptionID, vm.NodeResourceGroup, vm.VMScaleSet, vm.InstanceID); err != nil {
			return fmt.Errorf("failed to set node config for %s: %w", nn, err)
		}
	}
	return nil
}
