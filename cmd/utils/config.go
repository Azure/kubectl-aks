// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/spf13/viper"
)

const (
	currentNodeKey = "current-node"
	configFilename = "config.yaml"
)

var config = viper.New()

func init() {
	config.SetConfigFile(path.Join(configDir(), configFilename))
}

// ShowConfig prints the configuration to stdout
func ShowConfig() error {
	cfg, err := os.ReadFile(path.Join(configDir(), configFilename))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading config file: %w", err)
	}
	fmt.Fprintf(os.Stdout, "%s", cfg)
	return nil
}

// UseNodeConfig sets the current node to use in the configuration
func UseNodeConfig(targetNode string) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if !config.IsSet("nodes." + targetNode) {
		return fmt.Errorf("node %q not found", targetNode)
	}
	config.Set(currentNodeKey, targetNode)
	if err := config.WriteConfig(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// UnsetCurrentNodeConfig removes the current node from the configuration
func UnsetCurrentNodeConfig() error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if config.IsSet(currentNodeKey) {
		if err := deleteCurrentNodeConfig(); err != nil {
			return fmt.Errorf("deleting current node config: %w", err)
		}
		if err := config.WriteConfig(); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
	}
	return nil
}

// UnsetNodeConfig removes the node configuration
func UnsetNodeConfig(targetNode string) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if config.IsSet("nodes." + targetNode) {
		if err := deleteNodeConfig(targetNode); err != nil {
			return fmt.Errorf("deleting node config: %w", err)
		}
		if err := config.WriteConfig(); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
	}

	return nil
}

// UnsetAllConfig removes all the configuration
func UnsetAllConfig() error {
	if err := os.Remove(path.Join(configDir(), configFilename)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("deleting config file: %w", err)
	}
	return nil
}

// SetNodeConfig sets the node configuration based on the provided flags
func SetNodeConfig(targetNode string) error {
	if err := os.MkdirAll(configDir(), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	if err := loadConfig(); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	// delete the node config if it already exists to avoid mutually exclusive config
	if config.IsSet("nodes." + targetNode) {
		if err := deleteNodeConfig(targetNode); err != nil {
			return fmt.Errorf("deleting node config: %w", err)
		}
	}

	if node != "" {
		config.Set("nodes."+targetNode+".node", node)
	} else if resourceID != "" {
		config.Set("nodes."+targetNode+".id", resourceID)
	} else {
		config.Set("nodes."+targetNode+".subscription", subscriptionID)
		config.Set("nodes."+targetNode+".node-resource-group", nodeResourceGroup)
		config.Set("nodes."+targetNode+".vmss", vmss)
		config.Set("nodes."+targetNode+".instance-id", vmssInstanceID)
	}

	if err := config.WriteConfig(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

func deleteNodeConfig(targetNode string) error {
	return deleteConfig(func(settings map[string]interface{}) {
		delete(settings["nodes"].(map[string]interface{}), targetNode)
	})
}

func deleteCurrentNodeConfig() error {
	return deleteConfig(func(settings map[string]interface{}) {
		delete(settings, currentNodeKey)
	})
}

// deleteConfig, viper does not support deleting a config with key, so we get
// the underlying map, delete the key via deleteKey and then re-read the config
// https://github.com/spf13/viper/issues/632#issuecomment-869668629
func deleteConfig(deleteKey func(setting map[string]interface{})) error {
	settings := config.AllSettings()
	deleteKey(settings)
	data, err := json.MarshalIndent(settings, "", " ")
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}
	if err = config.ReadConfig(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("reading config: %w", err)
	}
	return nil
}

func loadConfig() error {
	if err := config.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading config: %w", err)
	}
	return nil
}

func loadCurrentNodeConfig() error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	currentNode := config.GetString(currentNodeKey)
	if currentNode != "" {
		config = config.Sub("nodes." + currentNode)
	}
	return nil
}

// configDir returns the directory where the cache file is stored.
// It will use the home directory if possible otherwise the current directory.
func configDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warn: getting home directory: %s\n", err)
		fmt.Fprintf(os.Stderr, "Warn: using current directory for config\n")
	}
	return path.Join(dir, ".kubectl-az")
}
