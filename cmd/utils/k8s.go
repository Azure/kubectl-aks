// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"context"
	"strings"

	"github.com/kinvolk/inspektor-gadget/pkg/k8sutil"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)

// GetNodeResourceID retrieve the Azure resource ID of a given node. In other
// words, the resource ID of the VM scale set instance. It returns format:
// /subscriptions/mySubID/resourceGroups/myRG/providers/myProvider/virtualMachineScaleSets/myVMSS/virtualMachines/myInsID
func GetNodeResourceID(ctx context.Context, nodeName string) (string, error) {
	client, err := k8sutil.NewClientsetFromConfigFlags(KubernetesConfigFlags)
	if err != nil {
		return "", err
	}

	nodeRes, err := client.CoreV1().Nodes().Get(ctx, nodeName, metaV1.GetOptions{})
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(nodeRes.Spec.ProviderID, "azure://"), nil
}
