// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

import (
	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
)

// Mode describes how a check executes on a node.
type Mode int

const (
	// ModeVerify runs a command once and interprets the result as pass/fail.
	ModeVerify Mode = iota
	// ModeTrace runs a command for a specified duration, collecting events.
	ModeTrace
)

// Result represents the outcome of a single check on one node.
type Result struct {
	Success bool
	Message string // human-readable one-line summary
	Details string // optional verbose output
}

// Check is the interface every check must implement.
// The framework handles runtime selection, node targeting, cluster fan-out,
// cobra wiring, and result reporting. Contributors only implement this.
type Check interface {
	// Name returns the kebab-case identifier used as the subcommand name.
	Name() string
	// Description returns a short description for --help.
	Description() string
	// Mode returns whether this is a verify (point-in-time) or trace (duration) check.
	Mode() Mode
	// Command returns the shell command(s) to execute on the node.
	// For trace checks, the framework substitutes {{.Duration}} with the
	// user-specified --duration value in seconds.
	Command() string
	// Parse interprets the runtime result and returns a check Result.
	Parse(res *pkgruntime.RunResult) (*Result, error)
}
