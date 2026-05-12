// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cmd

import (
	"github.com/Azure/kubectl-aks/cmd/utils/config"
)

// resolveRuntimeFromConfig reads runtime and debug-image from the config file
// if they were not explicitly set via CLI flags.
func resolveRuntimeFromConfig() {
	cfg := config.New()
	if err := cfg.ReadInConfig(); err != nil {
		return
	}

	// If the flag is still at its default, check the config file
	if runtimeFlag == RuntimeAzureAPI {
		if v := cfg.GetString(runtimeKey); v != "" {
			runtimeFlag = v
		}
	}
	if debugImage == "busybox:latest" {
		if v := cfg.GetString(debugImageKey); v != "" {
			debugImage = v
		}
	}
}
