// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/go-autorest/autorest/to"
)

type VirtualMachineScaleSetVM struct {
	SubscriptionID    string
	NodeResourceGroup string
	VMScaleSet        string
	InstanceID        string
}

// ParseVMSSResourceID extracts elements from a given VMSS resource ID with format:
// /subscriptions/mySubID/resourceGroups/myRG/providers/myProvider/virtualMachineScaleSets/myVMSS/virtualMachines/myInsID
func ParseVMSSResourceID(id string, vm *VirtualMachineScaleSetVM) error {
	const expectedItems int = 5

	// This allows us to make resource ID (--id) option not case sentitive
	id = strings.ToLower(id)

	// Required because fmt.Sscanf expects space-separated values
	idWithSpaces := strings.TrimSpace(strings.Replace(id, "/", " ", -1))

	// We don't need the provider but fmt.Sscanf does not support "%*s" operator
	// to read but prevent conversion. Therefore, read it and don't use it.
	var provider string

	n, err := fmt.Sscanf(idWithSpaces, "subscriptions %s resourcegroups %s providers %s virtualmachinescalesets %s virtualmachines %s",
		&vm.SubscriptionID, &vm.NodeResourceGroup, &provider, &vm.VMScaleSet, &vm.InstanceID)
	if err != nil {
		return fmt.Errorf("error parsing provider ID %s: %w", id, err)
	}
	if n != expectedItems {
		return fmt.Errorf("%d values retrieved while expecting %d when parsing id %s",
			n, expectedItems, id)
	}

	return nil
}

// VirtualMachineScaleSetVMFromConfig returns a VirtualMachineScaleSetVM object it assumes that the config is set and valid
func VirtualMachineScaleSetVMFromConfig() (*VirtualMachineScaleSetVM, error) {
	var vm VirtualMachineScaleSetVM
	if node != "" {
		var err error
		resourceID, err = GetNodeResourceID(context.TODO(), node)
		if err != nil {
			return nil, fmt.Errorf("retrieving Azure resource ID of node %s from API server: %w",
				node, err)
		}
		if err = ParseVMSSResourceID(resourceID, &vm); err != nil {
			return nil, fmt.Errorf("parsing Azure resource ID %s: %w", resourceID, err)
		}
	} else if resourceID != "" {
		if err := ParseVMSSResourceID(resourceID, &vm); err != nil {
			return nil, fmt.Errorf("parsing Azure resource ID %s: %w", resourceID, err)
		}
	} else {
		vm.SubscriptionID = subscriptionID
		vm.NodeResourceGroup = nodeResourceGroup
		vm.VMScaleSet = vmss
		vm.InstanceID = vmssInstanceID
	}

	return &vm, nil
}

func RunCommand(
	ctx context.Context,
	cred azcore.TokenCredential,
	vm *VirtualMachineScaleSetVM,
	command *string,
	verbose bool,
) (
	string,
	error,
) {
	const (
		commandID    = "RunShellScript"
		initialDelay = 15 * time.Second
		pollingFreq  = 2 * time.Second
	)

	client := armcompute.NewVirtualMachineScaleSetVMsClient(vm.SubscriptionID, cred, nil)

	script := []*string{command}
	runCommand := armcompute.RunCommandInput{
		CommandID: to.StringPtr(commandID),
		Script:    script,
	}

	if verbose {
		b, _ := json.MarshalIndent(vm, "", "  ")
		fmt.Printf("Command: %s\nVirtual Machine Scale Set VM:\n%s\n\n", *command, string(b))
	}

	poller, err := client.BeginRunCommand(ctx, vm.NodeResourceGroup,
		vm.VMScaleSet, vm.InstanceID, runCommand, nil)
	if err != nil {
		return "", fmt.Errorf("couldn't begin running command: %w", err)
	}

	fmt.Println("Running...")

	res, err := poller.PollUntilDone(ctx, pollingFreq)
	if err != nil {
		return "", fmt.Errorf("error polling command response: %w", err)
	}

	if verbose {
		b, _ := json.MarshalIndent(res, "", "  ")
		fmt.Printf("\nResponse:\n%s\n", string(b))
	}

	// TODO: Is it possible to have multiple values after using PollUntilDone()?
	if len(res.Value) == 0 || res.Value[0] == nil {
		return "", errors.New("no response received after command execution")
	}
	val := res.Value[0]

	// TODO: Isn't there a constant in the SDK to compare this?
	if to.String(val.Code) != "ProvisioningState/succeeded" {
		b, _ := json.MarshalIndent(res, "", "  ")
		return "", fmt.Errorf("command execution didn't succeed:\n%s", string(b))
	}

	// Expected format: "Enable succeeded: \n<text>"
	return strings.TrimPrefix(to.String(val.Message), "Enable succeeded: \n"), nil
}
