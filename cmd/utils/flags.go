// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"context"
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
func AddNodeFlags(command *cobra.Command, vm *VirtualMachineScaleSetVM) {
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

		if node != "" {
			if resourceID != "" {
				return errors.New("specify either --node or --id but not both")
			}

			var err error
			resourceID, err = GetNodeResourceID(context.TODO(), node)
			if err != nil {
				return fmt.Errorf("failed to retrieve Azure resource ID of node %s from API server: %w",
					node, err)
			}
		}

		if subscriptionID != "" && nodeResourceGroup != "" && vmScaleSet != "" && instanceID != "" {
			if resourceID != "" {
				return errors.New("do not provide VMMS instance information (--subscription, --node-resource-group, --vmss and --instance-id) when --node or --id were provided")
			}

			vm.SubscriptionID = subscriptionID
			vm.NodeResourceGroup = nodeResourceGroup
			vm.VMScaleSet = vmScaleSet
			vm.InstanceID = instanceID
		} else if resourceID != "" {
			if err := ParseVMSSResourceID(resourceID, vm); err != nil {
				return fmt.Errorf("failed to parse resource id: %w", err)
			}
		} else {
			return errors.New("specify either --node or --id or VMMS instance information (--subscription, --node-resource-group, --vmss and --instance-id)")
		}

		return nil
	}
}
