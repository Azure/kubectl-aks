// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"os"
	"path"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
)

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
