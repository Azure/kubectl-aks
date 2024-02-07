// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/Azure/kubectl-aks/cmd/utils/config"
	"github.com/kinvolk/inspektor-gadget/pkg/k8sutil"
	log "github.com/sirupsen/logrus"
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

type RunCommandResult struct {
	Stdout string
	Stderr string
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

// VirtualMachineScaleSetVMFromConfig returns a VirtualMachineScaleSetVM object
// it assumes that the config is set and valid
func VirtualMachineScaleSetVMFromConfig() (*VirtualMachineScaleSetVM, error) {
	var vm VirtualMachineScaleSetVM
	if node != "" {
		// Before trying to get the resource ID from the API server, verify if
		// the VMSS information of that node is already in the config file.
		config := config.New()
		if cc, ok := config.GetNodeConfig(node); ok {
			log.Debugf("Using VMSS information from config for node %s", node)

			vm.SubscriptionID = cc.GetString(SubscriptionIDKey)
			vm.NodeResourceGroup = cc.GetString(NodeResourceGroupKey)
			vm.VMScaleSet = cc.GetString(VMSSKey)
			vm.InstanceID = cc.GetString(VMSSInstanceIDKey)
			return &vm, nil
		}

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
	aksClient, err := armcontainerservice.NewManagedClustersClient(subID, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("creating AKS client: %w", err)
	}
	cluster, err := aksClient.Get(ctx, rg, clusterName, nil)
	if err != nil {
		return nil, fmt.Errorf("getting cluster: %w", err)
	}
	var nodePools []string
	vmssClient, err := armcompute.NewVirtualMachineScaleSetsClient(subID, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("creating VMSS client: %w", err)
	}
	nodePoolPager := vmssClient.NewListPager(to.String(cluster.Properties.NodeResourceGroup), nil)
	for nodePoolPager.More() {
		nextResult, err := nodePoolPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting next page of node pools: %w", err)
		}
		for _, np := range nextResult.Value {
			nodePools = append(nodePools, to.String(np.Name))
		}
	}
	vmssVMs := make(map[string]*VirtualMachineScaleSetVM)
	vmClient, err := armcompute.NewVirtualMachineScaleSetVMsClient(subID, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("creating VMSS VMs client: %w", err)
	}
	for _, np := range nodePools {
		instances, err := instancesForNodePool(ctx, vmClient, np, to.String(cluster.Properties.NodeResourceGroup))
		if err != nil {
			return nil, fmt.Errorf("getting instances for node pool %q: %w", np, err)
		}
		for _, instance := range instances {
			vmssVMs[instanceName(instance)] = &VirtualMachineScaleSetVM{
				SubscriptionID:    subID,
				VMScaleSet:        np,
				NodeResourceGroup: strings.ToLower(to.String(cluster.Properties.NodeResourceGroup)),
				InstanceID:        to.String(instance.InstanceID),
			}
		}
	}

	return vmssVMs, nil
}

func instancesForNodePool(ctx context.Context, vmClient *armcompute.VirtualMachineScaleSetVMsClient, pool, resourceGroup string) ([]*armcompute.VirtualMachineScaleSetVM, error) {
	var instances []*armcompute.VirtualMachineScaleSetVM
	pager := vmClient.NewListPager(resourceGroup, pool, nil)
	for pager.More() {
		nextPage, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		instances = append(instances, nextPage.Value...)
	}
	return instances, nil
}

// instanceName returns the instance name of the VMSS VM formatted as Kubernetes node name.
func instanceName(vm *armcompute.VirtualMachineScaleSetVM) string {
	if vm.Properties.OSProfile == nil || vm.Properties.OSProfile.ComputerName == nil {
		return to.String(vm.Name)
	}
	return strings.ToLower(to.String(vm.Properties.OSProfile.ComputerName))
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
	*RunCommandResult,
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

	client, err := armcompute.NewVirtualMachineScaleSetVMsClient(vm.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating VMSS VMs client: %w", err)
	}

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

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-s
		log.Warn("The requested command hasn't finished yet, hit 'Ctrl+C' again to exit anyway.")
		log.Warn("However, please notice the command will continue running in the node anyway, " +
			"and you will be unable to see the output or run another command until it finishes.")
		<-s
		os.Exit(1)
	}()

	DefaultSpinner.Start()
	DefaultSpinner.Suffix = " Running..."

	poller, err := client.BeginRunCommand(ctx, vm.NodeResourceGroup,
		vm.VMScaleSet, vm.InstanceID, runCommand, nil)
	if err != nil {
		DefaultSpinner.Stop()
		return nil, fmt.Errorf("begin running command: %w", err)
	}

	res, err := poller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: pollingFreq})
	DefaultSpinner.Stop()
	if err != nil {
		return nil, fmt.Errorf("polling command response: %w", err)
	}

	if verbose {
		b, _ := json.MarshalIndent(res, "", "  ")
		fmt.Printf("\nResponse:\n%s\n", string(b))
	}

	// TODO: Is it possible to have multiple values after using PollUntilDone()?
	if len(res.Value) == 0 || res.Value[0] == nil {
		return nil, errors.New("no response received after command execution")
	}
	val := res.Value[0]

	// TODO: Isn't there a constant in the SDK to compare this?
	if to.String(val.Code) != "ProvisioningState/succeeded" {
		b, _ := json.MarshalIndent(res, "", "  ")
		return nil, fmt.Errorf("command execution didn't succeed:\n%s", string(b))
	}

	result, err := parseRunCommandMessage(to.String(val.Message))
	if err != nil {
		return nil, err
	}
	if outputTruncate == OutputTruncateTail && result.isTruncated() {
		result.Stdout = fmt.Sprintf("%s... (truncated)\n", result.Stdout)
	}

	return result, nil
}

func parseRunCommandMessage(msg string) (*RunCommandResult, error) {
	// Expected format: "Enable succeeded: <text>"
	res := strings.TrimPrefix(msg, "Enable succeeded: ")

	// Extract stdout and stderr from response.
	// Expected format: "\n[stdout]\n<text>\n[stderr]\n<text>"
	split := regexp.MustCompile(`\n\[(stdout|stderr)\]\n`).Split(res, -1)
	if len(split) != 3 {
		return nil, fmt.Errorf("couldn't parse response message:\n%s", res)
	}
	return &RunCommandResult{
		Stdout: split[1],
		Stderr: split[2],
	}, nil
}

func (r *RunCommandResult) isTruncated() bool {
	return len(r.Stdout)+len(r.Stderr) >= BytesLimit
}
