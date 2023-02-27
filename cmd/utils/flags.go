// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

const (
	NodeKey              = "node"
	SubscriptionIDKey    = "subscription"
	NodeResourceGroupKey = "node-resource-group"
	VMSSKey              = "vmss"
	InstanceIDKey        = "instance-id"
	ResourceIDKey        = "id"
)

// We need package level variables to ensure that the viper flag binding works correctly.
// See: https://github.com/spf13/viper/issues/375#issuecomment-578552586
var (
	node              string
	subscriptionID    string
	nodeResourceGroup string
	vmScaleSet        string
	instanceID        string
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

// Every command that allows user to specify the node name has three options:
// (1) Provide the kubernetes node name
// (2) Provide the VMMS instance information (--subscription, --node-resource-group, --vmss and --instance-id)
// (3) Provide Resource ID (/subscriptions/mySubID/resourceGroups/myRG/providers/myProvider/virtualMachineScaleSets/myVMSS/virtualMachines/myInsID)
func AddNodeFlags(command *cobra.Command) {
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
		&vmScaleSet,
		VMSSKey, "",
		"",
		"Virtual machine scale set name.",
	)
	command.PersistentFlags().StringVarP(
		&instanceID,
		InstanceIDKey, "",
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
		if err := LoadCurrentInstanceConfig(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if err := BindEnvAndFlags(cmd); err != nil {
			return fmt.Errorf("failed to bind env and flags: %w", err)
		}

		// set the values from config with precedence:
		// (1) CLI flag
		// (2) environment variable
		// (3) config file
		node = GetConfig(NodeKey)
		subscriptionID = GetConfig(SubscriptionIDKey)
		nodeResourceGroup = GetConfig(NodeResourceGroupKey)
		vmScaleSet = GetConfig(VMSSKey)
		instanceID = GetConfig(InstanceIDKey)
		resourceID = GetConfig(ResourceIDKey)

		// validate the config
		var nodeSet, vmssSet, resourceIDSet bool
		if node != "" {
			nodeSet = true
		}
		if subscriptionID != "" && nodeResourceGroup != "" && vmScaleSet != "" && instanceID != "" {
			vmssSet = true
		}
		if resourceID != "" {
			resourceIDSet = true
		}
		if !nodeSet && !vmssSet && !resourceIDSet {
			return errors.New("specify either 'node' or 'id' or VMMS instance information ('subscription', 'node-resource-group', 'vmss' and 'instance-id')")
		} else if nodeSet {
			if vmssSet {
				return errors.New("specify either 'node' or VMMS instance information ('subscription', 'node-resource-group', 'vmss' and 'instance-id')")
			}
			if resourceIDSet {
				return errors.New("specify either 'node' or 'id'")
			}
		} else if vmssSet {
			if resourceIDSet {
				return errors.New("specify either VMMS instance information ('subscription', 'node-resource-group', 'vmss' and 'instance-id') or 'id'")
			}
		}

		return nil
	}
}
