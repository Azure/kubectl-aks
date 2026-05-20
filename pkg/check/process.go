// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

import (
	"fmt"
	"strings"

	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
)

func init() {
	Register(&processHealth{})
}

type processHealth struct{}

func (c *processHealth) Name() string { return "process-health" }
func (c *processHealth) Description() string {
	return "Check that critical node processes (kubelet, containerd) are running"
}
func (c *processHealth) Mode() Mode { return ModeVerify }

func (c *processHealth) Command() string {
	// Check critical services. Print "service:active" or "service:inactive/failed".
	return `for svc in kubelet containerd; do
  status=$(systemctl is-active "$svc" 2>/dev/null || echo "unknown")
  echo "$svc:$status"
done`
}

func (c *processHealth) Parse(res *pkgruntime.RunResult) (*Result, error) {
	lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
	var failed []string
	var statuses []string
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		svc, status := parts[0], strings.TrimSpace(parts[1])
		statuses = append(statuses, fmt.Sprintf("%s=%s", svc, status))
		if status != "active" {
			failed = append(failed, fmt.Sprintf("%s (%s)", svc, status))
		}
	}

	if len(failed) > 0 {
		return &Result{
			Success: false,
			Message: fmt.Sprintf("Critical processes not running: %s", strings.Join(failed, ", ")),
			Details: strings.Join(statuses, "\n"),
		}, nil
	}
	return &Result{
		Success: true,
		Message: fmt.Sprintf("All critical processes running: %s", strings.Join(statuses, ", ")),
	}, nil
}
