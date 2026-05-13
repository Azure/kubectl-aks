// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

const (
	currentClusterKey = "current-cluster"
	clustersKey       = "clusters"
)

// IsLegacyConfig returns true if the config uses the old format (top-level
// "nodes" without "clusters"). Callers should warn the user to re-import.
func (c *Config) IsLegacyConfig() bool {
	if err := c.ReadInConfig(); err != nil {
		return false
	}
	return c.IsSet("nodes") && !c.IsSet(clustersKey)
}

// UseClusterConfig sets the current cluster in the configuration.
func (c *Config) UseClusterConfig(clusterName string) error {
	if err := c.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading config: %w", err)
	}
	if !c.IsSet(clustersKey + "." + clusterName) {
		return fmt.Errorf("cluster %q not found", clusterName)
	}
	c.Set(currentClusterKey, clusterName)
	if err := c.WriteConfig(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// UnsetCurrentClusterConfig removes the current-cluster selection.
func (c *Config) UnsetCurrentClusterConfig() error {
	if err := c.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading config: %w", err)
	}
	if c.IsSet(currentClusterKey) {
		if err := c.deleteConfig(func(settings map[string]interface{}) {
			delete(settings, currentClusterKey)
		}); err != nil {
			return fmt.Errorf("deleting current cluster config: %w", err)
		}
		if err := c.WriteConfig(); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
	}
	return nil
}

// UnsetClusterConfig removes an entire cluster and its nodes from the config.
func (c *Config) UnsetClusterConfig(clusterName string) error {
	if err := c.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading config: %w", err)
	}
	if !c.IsSet(clustersKey + "." + clusterName) {
		return nil
	}

	// If the removed cluster was the current one, also unset current-cluster.
	if c.GetString(currentClusterKey) == clusterName {
		if err := c.deleteConfig(func(settings map[string]interface{}) {
			delete(settings, currentClusterKey)
		}); err != nil {
			return fmt.Errorf("deleting current cluster config: %w", err)
		}
	}

	if err := c.deleteConfig(func(settings map[string]interface{}) {
		clusters, ok := settings[clustersKey].(map[string]interface{})
		if ok {
			delete(clusters, clusterName)
		}
	}); err != nil {
		return fmt.Errorf("deleting cluster config: %w", err)
	}
	if err := c.WriteConfig(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// ListClusters returns the names of all clusters in the configuration.
func (c *Config) ListClusters() ([]string, error) {
	if err := c.ReadInConfig(); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	sub := c.Sub(clustersKey)
	if sub == nil {
		return nil, nil
	}
	settings := sub.AllSettings()
	names := make([]string, 0, len(settings))
	for k := range settings {
		names = append(names, k)
	}
	return names, nil
}

// CurrentClusterName returns the current cluster name, or "" if none is set.
func (c *Config) CurrentClusterName() string {
	if err := c.ReadInConfig(); err != nil {
		return ""
	}
	return c.GetString(currentClusterKey)
}

// SetClusterMetadata stores the cluster-level metadata (subscription,
// resource-group, cluster-name) under clusters.<clusterName>.
func (c *Config) SetClusterMetadata(clusterName, subscriptionID, resourceGroup string) error {
	if err := os.MkdirAll(Dir(), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	if err := c.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading config: %w", err)
	}
	prefix := clustersKey + "." + clusterName
	c.Set(prefix+".subscription", subscriptionID)
	c.Set(prefix+".resource-group", resourceGroup)
	c.Set(prefix+".cluster-name", clusterName)
	if err := c.WriteConfig(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// GetClusterNodeConfig returns the viper sub-tree for a node within a cluster.
func (c *Config) GetClusterNodeConfig(clusterName, nodeName string) (*Config, bool) {
	if err := c.ReadInConfig(); err != nil {
		return nil, false
	}
	key := clustersKey + "." + clusterName + ".nodes." + nodeName
	if v := c.Sub(key); v != nil {
		return &Config{Viper: v}, true
	}
	return nil, false
}

// ListClusterNodes returns all node names belonging to a cluster.
func (c *Config) ListClusterNodes(clusterName string) ([]string, error) {
	if err := c.ReadInConfig(); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	sub := c.Sub(clustersKey + "." + clusterName + ".nodes")
	if sub == nil {
		return nil, nil
	}
	settings := sub.AllSettings()
	names := make([]string, 0, len(settings))
	for k := range settings {
		names = append(names, k)
	}
	return names, nil
}

// ClearClusterNodes removes all nodes under a cluster while keeping the
// cluster metadata intact. This is used during import to sync the node list.
func (c *Config) ClearClusterNodes(clusterName string) error {
	if err := c.ReadInConfig(); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("reading config: %w", err)
	}
	nodesKey := clustersKey + "." + clusterName + ".nodes"
	if !c.IsSet(nodesKey) {
		return nil
	}
	if err := c.deleteConfig(func(settings map[string]interface{}) {
		clusters, ok := settings[clustersKey].(map[string]interface{})
		if !ok {
			return
		}
		cluster, ok := clusters[clusterName].(map[string]interface{})
		if !ok {
			return
		}
		delete(cluster, "nodes")
	}); err != nil {
		return fmt.Errorf("clearing cluster nodes: %w", err)
	}
	if err := c.WriteConfig(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
