// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
)

const (
	// DefaultTraceDuration is the default duration in seconds for trace checks.
	DefaultTraceDuration = 30
)

// NodeResult holds the outcome of running a check on a single node.
type NodeResult struct {
	NodeName string
	Result   *Result
	Err      error
}

// RunOnNode executes a check on a single node using the given runtime.
func RunOnNode(ctx context.Context, c Check, rt pkgruntime.Runtime, nodeName string, timeout int, duration int) (*NodeResult, error) {
	command := c.Command()

	// For trace checks, substitute the duration template.
	if c.Mode() == ModeTrace {
		command = strings.ReplaceAll(command, "{{.Duration}}", strconv.Itoa(duration))
		// Ensure the timeout is at least the trace duration + buffer.
		if timeout < duration+30 {
			timeout = duration + 30
		}
	}

	opts := &pkgruntime.RunOptions{
		NodeName: nodeName,
		Command:  command,
		Timeout:  timeout,
	}

	res, err := rt.RunCommand(ctx, opts)
	if err != nil {
		return &NodeResult{
			NodeName: nodeName,
			Err:      fmt.Errorf("running check %q: %w", c.Name(), err),
		}, nil
	}

	result, err := c.Parse(res)
	if err != nil {
		return &NodeResult{
			NodeName: nodeName,
			Err:      fmt.Errorf("parsing check %q result: %w", c.Name(), err),
		}, nil
	}

	return &NodeResult{
		NodeName: nodeName,
		Result:   result,
	}, nil
}

// RuntimeFactory creates a runtime for a given node name.
// This allows the runner to be decoupled from the runtime construction details.
type RuntimeFactory func(nodeName string) (pkgruntime.Runtime, error)

// RunOnNodes executes a check across multiple nodes in parallel.
func RunOnNodes(ctx context.Context, c Check, nodes []string, factory RuntimeFactory, timeout int, duration int) []NodeResult {
	results := make([]NodeResult, len(nodes))
	var wg sync.WaitGroup

	for i, nodeName := range nodes {
		wg.Add(1)
		go func(idx int, nn string) {
			defer wg.Done()

			rt, err := factory(nn)
			if err != nil {
				results[idx] = NodeResult{
					NodeName: nn,
					Err:      err,
				}
				return
			}

			nr, _ := RunOnNode(ctx, c, rt, nn, timeout, duration)
			results[idx] = *nr
		}(i, nodeName)
	}

	wg.Wait()
	return results
}

// FormatResults produces a consistent human-readable output and returns
// whether any check failed.
func FormatResults(results []NodeResult) (output string, hasFailure bool) {
	var b strings.Builder
	multiNode := len(results) > 1
	for i, r := range results {
		prefix := r.NodeName
		if prefix == "" {
			prefix = "(current node)"
		}
		if multiNode && i > 0 {
			fmt.Fprintln(&b)
		}
		if multiNode {
			fmt.Fprintf(&b, "=== %s ===\n", prefix)
		}
		if r.Err != nil {
			fmt.Fprintf(&b, "ERROR: %s\n", r.Err)
			hasFailure = true
		} else if r.Result.Success {
			fmt.Fprintf(&b, "✓ %s\n", r.Result.Message)
		} else {
			fmt.Fprintf(&b, "✗ %s\n", r.Result.Message)
			hasFailure = true
		}
		if r.Result != nil && r.Result.Details != "" {
			fmt.Fprintln(&b, r.Result.Details)
		}
	}
	return b.String(), hasFailure
}
