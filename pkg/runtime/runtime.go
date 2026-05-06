// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package runtime

import "context"

// Runtime abstracts command execution on a node.
type Runtime interface {
	RunCommand(ctx context.Context, opts *RunOptions) (*RunResult, error)
}

// RunOptions contains the options for running a command on a node.
type RunOptions struct {
	NodeName string
	Command  string
	Timeout  int
}

// RunResult contains the result of running a command on a node.
type RunResult struct {
	Stdout string
	Stderr string
}
