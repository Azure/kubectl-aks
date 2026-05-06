// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmss

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/Azure/kubectl-aks/cmd/utils"
	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
)

// Runtime executes commands on AKS nodes via the Azure VMSS RunCommand API.
type Runtime struct {
	Credential     azcore.TokenCredential
	VM             *utils.VirtualMachineScaleSetVM
	OutputTruncate utils.OutputTruncate
}

func (r *Runtime) RunCommand(ctx context.Context, opts *pkgruntime.RunOptions) (*pkgruntime.RunResult, error) {
	if r.Credential == nil {
		return nil, fmt.Errorf("credential is required for azure-api runtime")
	}
	if r.VM == nil {
		return nil, fmt.Errorf("VM information is required for azure-api runtime")
	}

	timeout := opts.Timeout
	command := opts.Command

	res, err := utils.RunCommand(ctx, r.Credential, r.VM, &command, &timeout, r.OutputTruncate)
	if err != nil {
		return nil, err
	}

	return &pkgruntime.RunResult{
		Stdout: res.Stdout,
		Stderr: res.Stderr,
	}, nil
}
