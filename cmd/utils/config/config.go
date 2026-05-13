// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package config

import (
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

type Config struct {
	*viper.Viper
	configPath string
}

func New() *Config {
	v := viper.New()
	configPath := path.Join(Dir(), configFilename)
	v.SetConfigFile(configPath)
	return &Config{Viper: v, configPath: configPath}
}

// Dir returns the directory where the config file is stored.
// It will use the home directory if possible otherwise the current directory.
func Dir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warn: getting home directory: %s\n", err)
		fmt.Fprintf(os.Stderr, "Warn: using current directory for config\n")
	}
	return path.Join(dir, ".kubectl-aks")
}

// ShowConfig prints the configuration to stdout
func (c *Config) ShowConfig() error {
	cfg, err := os.ReadFile(c.configPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading config file: %w", err)
	}
	fmt.Fprintf(os.Stdout, "%s", cfg)
	return nil
}

// UnsetAllConfig removes all the configuration
func (c *Config) UnsetAllConfig() error {
	if err := os.Remove(c.configPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("deleting config file: %w", err)
	}
	return nil
}

// CurrentConfig returns the current node configuration if it exists.
// It first checks cluster-aware config, then falls back to legacy.
func (c *Config) CurrentConfig() (*Config, bool) {
	if err := c.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, false
	}

	currentNode := c.GetString(currentNodeKey)
	if currentNode == "" {
		return nil, false
	}

	// Cluster-aware: search for node in current cluster first, then all clusters
	if c.IsSet(clustersKey) {
		currentCluster := c.GetString(currentClusterKey)
		if currentCluster != "" {
			key := clustersKey + "." + currentCluster + ".nodes." + currentNode
			if v := c.Sub(key); v != nil {
				return &Config{Viper: v}, true
			}
		}
		// Search all clusters
		clusters, _ := c.ListClusters()
		for _, cl := range clusters {
			key := clustersKey + "." + cl + ".nodes." + currentNode
			if v := c.Sub(key); v != nil {
				return &Config{Viper: v}, true
			}
		}
	}

	// Legacy fallback
	if v := c.Sub("nodes." + currentNode); v != nil {
		return &Config{Viper: v}, true
	}
	return nil, false
}

// CurrentNodeName returns the current node name from the configuration, if set.
func (c *Config) CurrentNodeName() string {
	if err := c.ReadInConfig(); err != nil {
		return ""
	}
	return c.GetString(currentNodeKey)
}

// GetNodeConfig returns the configuration for the given node if it exists.
// It searches cluster-aware config first, then legacy.
func (c *Config) GetNodeConfig(node string) (*Config, bool) {
	if err := c.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, false
	}

	// Cluster-aware: search current cluster first, then all clusters
	if c.IsSet(clustersKey) {
		currentCluster := c.GetString(currentClusterKey)
		if currentCluster != "" {
			key := clustersKey + "." + currentCluster + ".nodes." + node
			if v := c.Sub(key); v != nil {
				return &Config{Viper: v}, true
			}
		}
		clusters, _ := c.ListClusters()
		for _, cl := range clusters {
			key := clustersKey + "." + cl + ".nodes." + node
			if v := c.Sub(key); v != nil {
				return &Config{Viper: v}, true
			}
		}
	}

	// Legacy fallback
	if v := c.Sub("nodes." + node); v != nil {
		return &Config{Viper: v}, true
	}
	return nil, false
}
