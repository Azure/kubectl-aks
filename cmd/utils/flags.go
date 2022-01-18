// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
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
	var (
		node              string
		subscriptionID    string
		nodeResourceGroup string
		vmScaleSet        string
		instanceID        string
		resourceID        string
	)

	command.PersistentFlags().StringVarP(
		&node,
		"node", "",
		"",
		"Kubernetes node name.",
	)
	command.PersistentFlags().StringVarP(
		&subscriptionID,
		"subscription", "",
		"",
		"Subscription ID.",
	)
	command.PersistentFlags().StringVarP(
		&nodeResourceGroup,
		"node-resource-group", "",
		"",
		"Node resource group name.",
	)
	command.PersistentFlags().StringVarP(
		&vmScaleSet,
		"vmss", "",
		"",
		"Virtual machine scale set name.",
	)
	command.PersistentFlags().StringVarP(
		&instanceID,
		"instance-id", "",
		"",
		"VM scale set instance ID.",
	)
	command.PersistentFlags().StringVarP(
		&resourceID,
		"id", "",
		"",
		`Resource ID containing all information of the VMSS instance using format:
		e.g. /subscriptions/mySubID/resourceGroups/myRG/providers/myProvider/virtualMachineScaleSets/myVMSS/virtualMachines/myInsID.
		Notice it is not case sensitive.`,
	)

	command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
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
