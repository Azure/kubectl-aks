// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

// knownCloudNames maps Azure CLI cloud names (case-insensitive) to SDK cloud
// configurations. Multiple aliases are supported to match the conventions used
// by the Azure CLI and tools like kubelogin.
var knownCloudNames = map[string]cloud.Configuration{
	"azurecloud":             cloud.AzurePublic,
	"azurepublic":            cloud.AzurePublic,
	"azurepubliccloud":       cloud.AzurePublic,
	"azureusgovernment":      cloud.AzureGovernment,
	"azureusgovernmentcloud": cloud.AzureGovernment,
	"azurechinacloud":        cloud.AzureChina,
}

// GetCloudConfiguration returns the cloud.Configuration for the active Azure CLI cloud.
// It reads ~/.azure/config ([cloud] name) to determine the active cloud name and maps
// it to a known cloud.Configuration. If detection fails, it defaults to AzurePublic.
func GetCloudConfiguration() cloud.Configuration {
	azureConfigDir := azureConfigDir()

	cloudName, err := activeCloudName(azureConfigDir)
	if err != nil {
		log.Debugf("could not determine active cloud name, defaulting to AzurePublic: %v", err)
		return cloud.AzurePublic
	}

	if cfg, ok := knownCloudNames[strings.ToLower(cloudName)]; ok {
		log.Debugf("detected Azure cloud %q", cloudName)
		return cfg
	}

	// For custom clouds (e.g. Azure Stack), read the cloud's endpoints from
	// the Azure CLI profile.
	cfg, err := customCloudConfig(azureConfigDir, cloudName)
	if err != nil {
		log.Debugf("could not load custom cloud %q configuration, defaulting to AzurePublic: %v", cloudName, err)
		return cloud.AzurePublic
	}
	log.Debugf("loaded custom cloud configuration for %q", cloudName)
	return cfg
}

// ARMClientOptions returns arm.ClientOptions configured for the given cloud.
func ARMClientOptions(cfg cloud.Configuration) *arm.ClientOptions {
	return &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Cloud: cfg,
		},
	}
}

// azureConfigDir returns the path to the Azure CLI configuration directory.
func azureConfigDir() string {
	if dir := os.Getenv("AZURE_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return path.Join(home, ".azure")
}

// activeCloudName reads the Azure CLI config file (~/.azure/config) to determine the
// active cloud. The `az cloud set` command stores the active cloud name under the
// [cloud] section's "name" key.
func activeCloudName(configDir string) (string, error) {
	configPath := path.Join(configDir, "config")
	cfg, err := ini.Load(configPath)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", configPath, err)
	}

	section, err := cfg.GetSection("cloud")
	if err != nil {
		// No [cloud] section means the default cloud is in use.
		return "AzureCloud", nil
	}

	name := strings.TrimSpace(section.Key("name").String())
	if name == "" {
		return "AzureCloud", nil
	}
	return name, nil
}

// customCloudConfig reads cloud endpoint configuration for a non-standard cloud
// from the Azure CLI's clouds.config file.
func customCloudConfig(configDir, cloudName string) (cloud.Configuration, error) {
	cloudsPath := path.Join(configDir, "clouds.config")
	cfg, err := ini.Load(cloudsPath)
	if err != nil {
		return cloud.Configuration{}, fmt.Errorf("reading %s: %w", cloudsPath, err)
	}

	section, err := cfg.GetSection(cloudName)
	if err != nil {
		return cloud.Configuration{}, fmt.Errorf("cloud %q not found in %s", cloudName, cloudsPath)
	}

	endpoint := section.Key("endpoint_resource_manager").String()
	authority := section.Key("endpoint_active_directory").String()
	if endpoint == "" || authority == "" {
		return cloud.Configuration{}, fmt.Errorf("cloud %q is missing endpoint_resource_manager or endpoint_active_directory", cloudName)
	}

	return cloud.Configuration{
		ActiveDirectoryAuthorityHost: authority,
		Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
			cloud.ResourceManager: {
				Audience: normalizeEndpoint(endpoint),
				Endpoint: normalizeEndpoint(endpoint),
			},
		},
	}, nil
}

// normalizeEndpoint ensures the endpoint has a trailing slash for consistent comparison.
func normalizeEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}
	return endpoint
}
