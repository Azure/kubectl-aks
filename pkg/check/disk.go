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
	Register(&diskPressure{})
}

type diskPressure struct{}

func (c *diskPressure) Name() string { return "disk-pressure" }
func (c *diskPressure) Description() string {
	return "Check disk usage and inode exhaustion on the node"
}
func (c *diskPressure) Mode() Mode { return ModeVerify }

func (c *diskPressure) Command() string {
	// Output disk usage for /, and inode usage for /.
	// Format: "disk:<percent>" and "inode:<percent>"
	return `root_usage=$(df / | awk 'NR==2 {gsub(/%/,""); print $5}')
inode_usage=$(df -i / | awk 'NR==2 {gsub(/%/,""); print $5}')
echo "disk:${root_usage}"
echo "inode:${inode_usage}"`
}

func (c *diskPressure) Parse(res *pkgruntime.RunResult) (*Result, error) {
	const threshold = 85

	lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
	values := make(map[string]int)
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		v, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			continue
		}
		values[parts[0]] = v
	}

	var issues []string
	if v, ok := values["disk"]; ok && v >= threshold {
		issues = append(issues, fmt.Sprintf("root disk %d%% used", v))
	}
	if v, ok := values["inode"]; ok && v >= threshold {
		issues = append(issues, fmt.Sprintf("inodes %d%% used", v))
	}

	if len(issues) > 0 {
		return &Result{
			Success: false,
			Message: fmt.Sprintf("Disk pressure: %s", strings.Join(issues, "; ")),
			Details: res.Stdout,
		}, nil
	}

	diskPct := values["disk"]
	inodePct := values["inode"]
	return &Result{
		Success: true,
		Message: fmt.Sprintf("Disk OK: root %d%% used, inodes %d%% used", diskPct, inodePct),
	}, nil
}
