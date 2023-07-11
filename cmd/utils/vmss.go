// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/kinvolk/inspektor-gadget/pkg/k8sutil"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OutputTruncate int

const (
	OutputTruncateHead OutputTruncate = iota
	OutputTruncateTail

	BytesLimit                        = 4096
	DefaultRunCommandTimeoutInSeconds = 300
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
		return fmt.Errorf("error parsing provider ID %q: %w", id, err)
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

func VirtualMachineScaleSetVMsViaKubeconfig() (map[string]*VirtualMachineScaleSetVM, error) {
	clientset, err := k8sutil.NewClientsetFromConfigFlags(KubernetesConfigFlags)
	if err != nil {
		return nil, fmt.Errorf("creating Kubernetes client: %w", err)
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}

	vmssVMs := make(map[string]*VirtualMachineScaleSetVM)
	if len(nodes.Items) > 0 {
		for _, n := range nodes.Items {
			var vm VirtualMachineScaleSetVM
			if !strings.HasPrefix(n.Spec.ProviderID, "azure://") {
				return nil, fmt.Errorf("node=%q doesn't seem to be an Azure VMSS VM", n.Name)
			}
			if err = ParseVMSSResourceID(strings.TrimPrefix(n.Spec.ProviderID, "azure://"), &vm); err != nil {
				return nil, fmt.Errorf("parsing Azure resource ID %q: %w", n.Spec.ProviderID, err)
			}
			vmssVMs[n.Name] = &vm
		}
	}
	return vmssVMs, nil
}

func VirtualMachineScaleSetVMsViaAzureAPI(subID, rg, clusterName string) (map[string]*VirtualMachineScaleSetVM, error) {
	creds, err := GetCredentials()
	if err != nil {
		return nil, fmt.Errorf("getting credentials: %w", err)
	}

	ctx := context.Background()
	aksClient := armcontainerservice.NewManagedClustersClient(subID, creds, nil)
	cluster, err := aksClient.Get(ctx, rg, clusterName, nil)
	if err != nil {
		return nil, fmt.Errorf("getting cluster: %w", err)
	}
	var nodePools []string
	vmssClient := armcompute.NewVirtualMachineScaleSetsClient(subID, creds, nil)
	nodePoolPager := vmssClient.List(to.String(cluster.Properties.NodeResourceGroup), nil)
	for nodePoolPager.NextPage(ctx) {
		for _, np := range nodePoolPager.PageResponse().Value {
			nodePools = append(nodePools, to.String(np.Name))
		}
	}
	if err = nodePoolPager.Err(); err != nil {
		return nil, fmt.Errorf("getting node pools: %w", err)
	}
	vmssVMs := make(map[string]*VirtualMachineScaleSetVM)
	vmClient := armcompute.NewVirtualMachineScaleSetVMsClient(subID, creds, nil)
	for _, np := range nodePools {
		instances, err := instancesForNodePool(vmClient, np, to.String(cluster.Properties.NodeResourceGroup))
		if err != nil {
			return nil, fmt.Errorf("getting instances for node pool %q: %w", np, err)
		}
		for _, instance := range instances {
			vmssVMs[instanceName(np, to.String(instance.InstanceID))] = &VirtualMachineScaleSetVM{
				SubscriptionID:    subID,
				VMScaleSet:        np,
				NodeResourceGroup: strings.ToLower(to.String(cluster.Properties.NodeResourceGroup)),
				InstanceID:        to.String(instance.InstanceID),
			}
		}
	}

	return vmssVMs, nil
}

func instancesForNodePool(vmClient *armcompute.VirtualMachineScaleSetVMsClient, pool, resourceGroup string) ([]*armcompute.VirtualMachineScaleSetVM, error) {
	var instances []*armcompute.VirtualMachineScaleSetVM
	pager := vmClient.List(resourceGroup, pool, nil)
	for pager.NextPage(context.Background()) {
		instances = append(instances, pager.PageResponse().Value...)
	}
	if err := pager.Err(); err != nil {
		return nil, err
	}
	return instances, nil
}

// instanceName returns the instance name of the VMSS VM formatted as Kubernetes node name.
func instanceName(vmss string, instanceID string) string {
	id, err := strconv.Atoi(instanceID)
	if err != nil {
		return fmt.Sprintf("%s_%s", vmss, instanceID)
	}
	return fmt.Sprintf("%s%06x", vmss, id)
}

func RunCommand(
	ctx context.Context,
	cred azcore.TokenCredential,
	vm *VirtualMachineScaleSetVM,
	command *string,
	verbose bool,
	timeout *int,
	outputTruncate OutputTruncate,
) (
	string,
	error,
) {
	const (
		commandID    = "RunShellScript"
		initialDelay = 15 * time.Second
		pollingFreq  = 2 * time.Second
	)

	if timeout == nil {
		timeout = to.IntPtr(DefaultRunCommandTimeoutInSeconds)
	}

	client := armcompute.NewVirtualMachineScaleSetVMsClient(vm.SubscriptionID, cred, nil)

	// By default, the Azure API limits the output to the last 4,096 bytes. See
	// https://learn.microsoft.com/en-us/azure/virtual-machines/linux/run-command#restrictions.
	if outputTruncate == OutputTruncateTail {
		*command = fmt.Sprintf("%s | head -c %d", *command, BytesLimit)
	}

	script := []*string{to.StringPtr(fmt.Sprintf("timeout %d sh -c '%s'", *timeout, *command))}
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
