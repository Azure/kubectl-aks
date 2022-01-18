// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// Further details about authentication:
// https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/azidentity
func GetCredentials() (*azidentity.ChainedTokenCredential, error) {
	azCLI, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating default authentication chain: %w", err)
	}

	// Fallback if users didn't get already authenticated using the Azure CLI
	inBrowser, err := azidentity.NewInteractiveBrowserCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating interactive authentication chain: %w", err)
	}

	// Methods will be tried in that specific order: (1) Azure CLI (2) Interactive
	chain, err := azidentity.NewChainedTokenCredential([]azcore.TokenCredential{azCLI, inBrowser}, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating credential chain: %w", err)
	}

	return chain, nil
}
