// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Azure/kubectl-aks/cmd/utils/config"
)

const (
	NodeKey              = "node"
	SubscriptionIDKey    = "subscription"
	ResourceGroupKey     = "resource-group"
	ClusterNameKey       = "cluster-name"
	NodeResourceGroupKey = "node-resource-group"
	VMSSKey              = "vmss"
	VMSSInstanceIDKey    = "instance-id"
	ResourceIDKey        = "id"
)

// We need package level variables to ensure that the viper flag binding works correctly.
// See: https://github.com/spf13/viper/issues/375#issuecomment-578552586
var (
	node              string
	subscriptionID    string
	nodeResourceGroup string
	vmss              string
	vmssInstanceID    string
	resourceID        string
)

// GetNodeName returns the current node name from flags/config.
func GetNodeName() string {
	return node
}

// CommonFlags contains CLI flags common for all subcommands
type CommonFlags struct {
	Verbose bool
}

func AddCommonFlags(command *cobra.Command, flags *CommonFlags) {
	command.PersistentFlags().BoolVarP(
		&flags.Verbose,
		"verbose", "v",
		false,
		"Verbose output.",
	)
}

// AddNodeFlags adds node flags and binds them to config/environment variables
// Every command that allows user to specify the node name has three options:
// (1) Provide the kubernetes node name
// (2) Provide the VMMS instance information (--subscription, --node-resource-group, --vmss and --instance-id)
// (3) Provide Resource ID (/subscriptions/mySubID/resourceGroups/myRG/providers/myProvider/virtualMachineScaleSets/myVMSS/virtualMachines/myInsID)
func AddNodeFlags(command *cobra.Command) {
	addNodeFlags(command, false)
}

// AddNodeFlagsOnly adds node flags without binding config/environment variables
func AddNodeFlagsOnly(command *cobra.Command) {
	addNodeFlags(command, true)
}

func addNodeFlags(command *cobra.Command, useFlagsOnly bool) {
	command.PersistentFlags().StringVarP(
		&node,
		NodeKey, "",
		"",
		"Kubernetes node name.",
	)
	command.PersistentFlags().StringVarP(
		&subscriptionID,
		SubscriptionIDKey, "",
		"",
		"Subscription ID.",
	)
	command.PersistentFlags().StringVarP(
		&nodeResourceGroup,
		NodeResourceGroupKey, "",
		"",
		"Node resource group name.",
	)
	command.PersistentFlags().StringVarP(
		&vmss,
		VMSSKey, "",
		"",
		"Virtual machine scale set name.",
	)
	command.PersistentFlags().StringVarP(
		&vmssInstanceID,
		VMSSInstanceIDKey, "",
		"",
		"VM scale set instance ID.",
	)
	command.PersistentFlags().StringVarP(
		&resourceID,
		ResourceIDKey, "",
		"",
		`Resource ID containing all information of the VMSS instance using format:
		e.g. /subscriptions/mySubID/resourceGroups/myRG/providers/myProvider/virtualMachineScaleSets/myVMSS/virtualMachines/myInsID.
		Notice it is not case sensitive.`,
	)

	command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// If node or resource ID is set, we don't need to read the config file
		// nor the environment variables because the CLI flags have precedence.
		if !useFlagsOnly && node == "" && resourceID == "" {
			cfg := config.New()

			// If a current node is set in config, remember the name
			currentNodeName := cfg.CurrentNodeName()

			if cc, ok := cfg.CurrentConfig(); ok {
				cfg = cc
			}
			// bind environment variables
			cfg.AutomaticEnv()
			cfg.SetEnvPrefix("kubectl_aks")
			cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

			// bind CLI flags
			if err := cfg.BindPFlags(cmd.PersistentFlags()); err != nil {
				return fmt.Errorf("binding flags: %w", err)
			}

			// set the values with precedence:
			// (1) CLI flag
			// (2) environment variable
			// (3) config file
			node = cfg.GetString(NodeKey)
			subscriptionID = cfg.GetString(SubscriptionIDKey)
			nodeResourceGroup = cfg.GetString(NodeResourceGroupKey)
			vmss = cfg.GetString(VMSSKey)
			vmssInstanceID = cfg.GetString(VMSSInstanceIDKey)
			resourceID = cfg.GetString(ResourceIDKey)

			// If node is still empty but we have a current node from config,
			// use it (needed for kube-api runtime which only requires node name)
			if node == "" && currentNodeName != "" {
				node = currentNodeName
			}
		}

		// validate the parameters
		var nodeSet, vmssInfoSet, resourceIDSet bool
		if node != "" {
			nodeSet = true
		}
		if subscriptionID != "" && nodeResourceGroup != "" && vmss != "" && vmssInstanceID != "" {
			vmssInfoSet = true
		}
		if resourceID != "" {
			resourceIDSet = true
		}
		if !nodeSet && !vmssInfoSet && !resourceIDSet {
			if subscriptionID != "" || nodeResourceGroup != "" || vmss != "" || vmssInstanceID != "" {
				return errors.New("specify complete VMMS instance information ('subscription', 'node-resource-group', 'vmss' and 'instance-id')")
			}
			return errors.New("specify either 'node' or 'id' or VMMS instance information ('subscription', 'node-resource-group', 'vmss' and 'instance-id')")
		} else if nodeSet && vmssInfoSet {
			// Both node name and VMSS info available (e.g., from config).
			// This is valid — the runtime will decide which to use.
			if resourceIDSet {
				return errors.New("specify either 'node' or 'id'")
			}
		} else if nodeSet {
			if resourceIDSet {
				return errors.New("specify either 'node' or 'id'")
			}
		} else if vmssInfoSet {
			if resourceIDSet {
				return errors.New("specify either VMMS instance information ('subscription', 'node-resource-group', 'vmss' and 'instance-id') or 'id'")
			}
		}

		return nil
	}
}
