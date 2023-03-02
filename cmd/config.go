// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Azure/kubectl-az/cmd/utils"
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

func init() {
	utils.AddCommonFlags(configCmd, &commonFlags)
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(showConfigCmd, useNodeCmd, unsetCurrentNodeCmd, unsetNodeCmd, unsetAllCmd, setNodeCmd)
	utils.AddNodeFlagsOnly(setNodeCmd)
}

func showConfigCmdRun(cmd *cobra.Command, args []string) error {
	return utils.ShowConfig()
}

func useNodeCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s <node name>", cmd.CommandPath())
	}
	return utils.UseNodeConfig(args[0])
}

func unsetCurrentNodeCmdRun(cmd *cobra.Command, args []string) error {
	return utils.UnsetCurrentNodeConfig()
}

func unsetNodeCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s <node name>", cmd.CommandPath())
	}
	return utils.UnsetNodeConfig(args[0])
}

func unsetAllCmdRun(cmd *cobra.Command, args []string) error {
	return utils.UnsetAllConfig()
}

func setNodeCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s <node name>", cmd.CommandPath())
	}
	return utils.SetNodeConfig(args[0])
}
