// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

type userAgentTestCredential struct{}

func (userAgentTestCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "test-token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

type userAgentTestTransport func(*http.Request) (*http.Response, error)

func (f userAgentTestTransport) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestARMClientOptions_RunCommandUserAgent(t *testing.T) {
	for _, tt := range []struct {
		name  string
		cloud cloud.Configuration
	}{
		{"public", cloud.AzurePublic},
		{"government", cloud.AzureGovernment},
		{"china", cloud.AzureChina},
	} {
		t.Run(tt.name, func(t *testing.T) {
			opts := ARMClientOptions(tt.cloud)
			opts.DisableRPRegistration = true
			requests := 0
			opts.Transport = userAgentTestTransport(func(req *http.Request) (*http.Response, error) {
				requests++
				if got := req.Header.Get("User-Agent"); !strings.HasPrefix(got, "kubectl-aks azsdk-go-armcompute/") {
					t.Errorf("User-Agent = %q, want kubectl-aks followed by SDK identifier", got)
				}
				endpoint := tt.cloud.Services[cloud.ResourceManager].Endpoint
				if !strings.HasPrefix(req.URL.String(), endpoint) {
					t.Errorf("request URL = %q, want cloud endpoint %q", req.URL, endpoint)
				}
				if req.Method != http.MethodPost || !strings.HasSuffix(req.URL.Path, "/runCommand") {
					t.Errorf("unexpected request: %s %s", req.Method, req.URL)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"value":[]}`)),
					Request:    req,
				}, nil
			})
			client, err := armcompute.NewVirtualMachineScaleSetVMsClient("subscription", userAgentTestCredential{}, opts)
			if err != nil {
				t.Fatal(err)
			}
			commandID := "RunShellScript"
			_, err = client.BeginRunCommand(context.Background(), "node-rg", "vmss", "0",
				armcompute.RunCommandInput{CommandID: &commandID}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Errorf("sent %d requests, want 1", requests)
			}
		})
	}
}

func TestActiveCloudName(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		want       string
	}{
		{
			name:       "AzureUSGovernment",
			configData: "[cloud]\nname = AzureUSGovernment\n",
			want:       "AzureUSGovernment",
		},
		{
			name:       "AzureCloud explicit",
			configData: "[cloud]\nname = AzureCloud\n",
			want:       "AzureCloud",
		},
		{
			name:       "no cloud section defaults to AzureCloud",
			configData: "[core]\noutput = table\n",
			want:       "AzureCloud",
		},
		{
			name:       "empty name defaults to AzureCloud",
			configData: "[cloud]\nname = \n",
			want:       "AzureCloud",
		},
		{
			name:       "AzureChinaCloud",
			configData: "[cloud]\nname = AzureChinaCloud\n",
			want:       "AzureChinaCloud",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(path.Join(dir, "config"), []byte(tt.configData), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := activeCloudName(dir)
			if err != nil {
				t.Fatalf("activeCloudName() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("activeCloudName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestActiveCloudName_MissingFile(t *testing.T) {
	_, err := activeCloudName(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestGetCloudConfiguration_KnownClouds(t *testing.T) {
	tests := []struct {
		cloudName string
		want      cloud.Configuration
	}{
		{"AzureCloud", cloud.AzurePublic},
		{"AzureUSGovernment", cloud.AzureGovernment},
		{"AzureChinaCloud", cloud.AzureChina},
	}

	for _, tt := range tests {
		t.Run(tt.cloudName, func(t *testing.T) {
			dir := t.TempDir()
			data := "[cloud]\nname = " + tt.cloudName + "\n"
			if err := os.WriteFile(path.Join(dir, "config"), []byte(data), 0o600); err != nil {
				t.Fatal(err)
			}

			// Override AZURE_CONFIG_DIR to point to our test directory.
			t.Setenv("AZURE_CONFIG_DIR", dir)

			got := GetCloudConfiguration()
			if got.ActiveDirectoryAuthorityHost != tt.want.ActiveDirectoryAuthorityHost {
				t.Errorf("ActiveDirectoryAuthorityHost = %q, want %q",
					got.ActiveDirectoryAuthorityHost, tt.want.ActiveDirectoryAuthorityHost)
			}
		})
	}
}
