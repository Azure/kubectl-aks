// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

// registry holds all registered checks.
var registry []Check

// Register adds a check to the global registry.
// Typically called from init() in each check file.
func Register(c Check) {
	registry = append(registry, c)
}

// All returns all registered checks.
func All() []Check {
	return registry
}

// ByMode returns checks filtered by mode.
func ByMode(m Mode) []Check {
	var out []Check
	for _, c := range registry {
		if c.Mode() == m {
			out = append(out, c)
		}
	}
	return out
}
