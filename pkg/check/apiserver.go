// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

import (
	"fmt"
	"strconv"
	"strings"

	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
)

func init() {
	Register(&apiserverConnectivity{})
}

type apiserverConnectivity struct{}

func (c *apiserverConnectivity) Name() string { return "apiserver-connectivity" }
func (c *apiserverConnectivity) Description() string {
	return "Check connectivity between the node and the Kubernetes API Server"
}
func (c *apiserverConnectivity) Mode() Mode { return ModeVerify }

func (c *apiserverConnectivity) Command() string {
	return `KUBECTL=$(command -v kubectl || echo kubectl); [ -x /opt/bin/kubectl ] && KUBECTL=/opt/bin/kubectl; $KUBECTL --kubeconfig /var/lib/kubelet/kubeconfig version > /dev/null; echo -n $?`
}

func (c *apiserverConnectivity) Parse(res *pkgruntime.RunResult) (*Result, error) {
	ret, err := strconv.Atoi(strings.TrimSpace(res.Stdout))
	if err != nil {
		return nil, fmt.Errorf("couldn't parse stdout of response message: %s", res.Stdout)
	}
	if ret != 0 {
		return &Result{
			Success: false,
			Message: fmt.Sprintf("Connectivity check: failed (exit code %d)", ret),
			Details: res.Stderr,
		}, nil
	}
	return &Result{
		Success: true,
		Message: "Connectivity check: succeeded",
	}, nil
}
