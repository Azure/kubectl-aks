// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"

	"github.com/Azure/kubectl-aks/cmd/utils/config"
)

// https://github.com/Azure/azure-sdk-for-go/blob/sdk/azidentity/v0.13.0/sdk/azidentity/azidentity.go#L25
const (
	organizationsTenantID   = "organizations"
	developerSignOnClientID = "04b07795-8ddb-461a-bbee-02f9e1bf7b46"
)

// GetCredentials returns a credential chain that will try to authenticate
// using the Azure CLI and then using the interactive browser.
// Further details about authentication:
// https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/azidentity
func GetCredentials() (*azidentity.ChainedTokenCredential, error) {
	azCLI, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating default authentication chain: %w", err)
	}

	// Fallback if users didn't get already authenticated using the Azure CLI
	inBrowser, err := newCachedInteractiveBrowserCredential()
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

// cachedInteractiveBrowserCredential is a credential that uses the interactive browser to authenticate and caches the token.
// TODO: This is a workaround until the azidentity package supports caching, https://github.com/Azure/azure-sdk-for-go/issues/16643.
type cachedInteractiveBrowserCredential struct {
	client public.Client
}

func newCachedInteractiveBrowserCredential() (*cachedInteractiveBrowserCredential, error) {
	file := path.Join(config.Dir(), "token-cache.json")
	if err := os.MkdirAll(path.Dir(file), 0o700); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	client, err := public.New(developerSignOnClientID,
		public.WithCache(&tokenCache{file: file}),
		public.WithAuthority(runtime.JoinPaths(string(azidentity.AzurePublicCloud), organizationsTenantID)),
	)
	if err != nil {
		return nil, fmt.Errorf("creating public client: %w", err)
	}

	return &cachedInteractiveBrowserCredential{client: client}, nil
}

// GetToken implements the azcore.TokenCredential interface on cachedInteractiveBrowserCredential.
func (c *cachedInteractiveBrowserCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (*azcore.AccessToken, error) {
	// TODO: may be this can be improved with https://github.com/Azure/kubectl-aks/issues/11
	var account public.Account
	if len(c.client.Accounts()) > 0 {
		account = c.client.Accounts()[len(c.client.Accounts())-1]
	}
	result, err := c.client.AcquireTokenSilent(ctx, options.Scopes, public.WithSilentAccount(account))
	if err != nil {
		result, err = c.client.AcquireTokenInteractive(ctx, options.Scopes)
		if err != nil {
			return nil, fmt.Errorf("acquiring interactive token: %w", err)
		}
	}
	return &azcore.AccessToken{Token: result.AccessToken, ExpiresOn: result.ExpiresOn}, nil
}

// tokenCache implements basic file based cache.ExportReplace to be used with the public.Client.
type tokenCache struct {
	file string
}

func (t *tokenCache) Replace(cache cache.Unmarshaler, key string) {
	data, err := os.ReadFile(t.file)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "Warn: reading token cache: %s\n", err)
	}
	err = cache.Unmarshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warn: unmarshaling token cache: %s\n", err)
	}
}

func (t *tokenCache) Export(cache cache.Marshaler, key string) {
	data, err := cache.Marshal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warn: marshaling token cache: %s\n", err)
	}
	var indentedData bytes.Buffer
	if err = json.Indent(&indentedData, data, "", "  "); err == nil {
		data = indentedData.Bytes()
	}
	err = os.WriteFile(t.file, data, 0o600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warn: writing token cache: %s\n", err)
	}
}
