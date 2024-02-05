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

type config struct {
	*viper.Viper
	configPath string
}

func New() *config {
	v := viper.New()
	configPath := path.Join(Dir(), configFilename)
	v.SetConfigFile(configPath)
	return &config{Viper: v, configPath: configPath}
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
func (c *config) ShowConfig() error {
	cfg, err := os.ReadFile(c.configPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading config file: %w", err)
	}
	fmt.Fprintf(os.Stdout, "%s", cfg)
	return nil
}

// UnsetAllConfig removes all the configuration
func (c *config) UnsetAllConfig() error {
	if err := os.Remove(c.configPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("deleting config file: %w", err)
	}
	return nil
}

// CurrentConfig returns the current node configuration if it exists
func (c *config) CurrentConfig() (*config, bool) {
	if err := c.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, false
	}
	currentNode := c.GetString(currentNodeKey)
	if currentNode == "" {
		return nil, false
	}
	return &config{Viper: c.Sub("nodes." + currentNode)}, true
}

// GetNodeConfig returns the configuration for the given node if it exists
func (c *config) GetNodeConfig(node string) (*config, bool) {
	if err := c.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, false
	}
	if v := c.Sub("nodes." + node); v != nil {
		return &config{Viper: v}, true
	}
	return nil, false
}
