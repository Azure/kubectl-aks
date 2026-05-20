// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

import (
	"fmt"
	"strings"

	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
)

func init() {
	Register(&oomEvents{})
}

type oomEvents struct{}

func (c *oomEvents) Name() string        { return "oom-events" }
func (c *oomEvents) Description() string { return "Check for recent OOM kill events on the node" }
func (c *oomEvents) Mode() Mode          { return ModeVerify }

func (c *oomEvents) Command() string {
	// Search dmesg and journald for OOM kills in the last hour.
	return `dmesg --time-format iso 2>/dev/null | grep -i "oom-kill\|out of memory" | tail -20 || journalctl -k --since "1 hour ago" --no-pager 2>/dev/null | grep -i "oom-kill\|out of memory" | tail -20; echo "EXIT:$?"`
}

func (c *oomEvents) Parse(res *pkgruntime.RunResult) (*Result, error) {
	stdout := strings.TrimSpace(res.Stdout)
	// Remove the EXIT: trailer
	lines := strings.Split(stdout, "\n")
	var oomLines []string
	for _, line := range lines {
		if strings.HasPrefix(line, "EXIT:") {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			oomLines = append(oomLines, trimmed)
		}
	}

	if len(oomLines) == 0 {
		return &Result{
			Success: true,
			Message: "No recent OOM kill events detected",
		}, nil
	}

	return &Result{
		Success: false,
		Message: fmt.Sprintf("Found %d OOM kill event(s)", len(oomLines)),
		Details: strings.Join(oomLines, "\n"),
	}, nil
}
