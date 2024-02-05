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
			config := config.New()

			if cc, ok := config.CurrentConfig(); ok {
				config = cc
			}
			// bind environment variables
			config.AutomaticEnv()
			config.SetEnvPrefix("kubectl_aks")
			config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

			// bind CLI flags
			if err := config.BindPFlags(cmd.PersistentFlags()); err != nil {
				return fmt.Errorf("binding flags: %w", err)
			}

			// set the values with precedence:
			// (1) CLI flag
			// (2) environment variable
			// (3) config file
			node = config.GetString(NodeKey)
			subscriptionID = config.GetString(SubscriptionIDKey)
			nodeResourceGroup = config.GetString(NodeResourceGroupKey)
			vmss = config.GetString(VMSSKey)
			vmssInstanceID = config.GetString(VMSSInstanceIDKey)
			resourceID = config.GetString(ResourceIDKey)
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
		} else if nodeSet {
			if vmssInfoSet {
				return errors.New("specify either 'node' or VMMS instance information ('subscription', 'node-resource-group', 'vmss' and 'instance-id')")
			}
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
