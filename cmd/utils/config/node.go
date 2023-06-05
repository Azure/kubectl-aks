// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// UseNodeConfig sets the current node to use in the configuration
func (c *config) UseNodeConfig(targetNode string) error {
	if err := c.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading config: %w", err)
	}
	if !c.IsSet("nodes." + targetNode) {
		return fmt.Errorf("node %q not found", targetNode)
	}
	c.Set(currentNodeKey, targetNode)
	if err := c.WriteConfig(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// UnsetCurrentNodeConfig removes the current node from the configuration
func (c *config) UnsetCurrentNodeConfig() error {
	if err := c.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading config: %w", err)
	}

	if c.IsSet(currentNodeKey) {
		if err := c.deleteCurrentNodeConfig(); err != nil {
			return fmt.Errorf("deleting current node config: %w", err)
		}
		if err := c.WriteConfig(); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
	}
	return nil
}

// UnsetNodeConfig removes the node configuration
func (c *config) UnsetNodeConfig(targetNode string) error {
	if err := c.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading config: %w", err)
	}

	if c.IsSet("nodes." + targetNode) {
		if err := c.deleteNodeConfig(targetNode); err != nil {
			return fmt.Errorf("deleting node config: %w", err)
		}
		if err := c.WriteConfig(); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
	}

	return nil
}

// SetNodeConfigWithNodeFlag sets the node configuration with based on node flag
func (c *config) SetNodeConfigWithNodeFlag(nodeName, nodeFlag string) error {
	return c.setNodeConfig(nodeName, nodeFlag, "", "", "", "", "")
}

// SetNodeConfigWithResourceIDFlag sets the node configuration with based on resource ID flag
func (c *config) SetNodeConfigWithResourceIDFlag(nodeName, resourceIDFlag string) error {
	return c.setNodeConfig(nodeName, "", resourceIDFlag, "", "", "", "")
}

// SetNodeConfigWithVMSSInfoFlag sets the node configuration with vmss info flags
func (c *config) SetNodeConfigWithVMSSInfoFlag(nodeName, subscriptionIDFlag, nodeResourceGroupFlag, vmssFlag, instanceIDFlag string) error {
	return c.setNodeConfig(nodeName, "", "", subscriptionIDFlag, nodeResourceGroupFlag, vmssFlag, instanceIDFlag)
}

func (c *config) setNodeConfig(nodeName, nodeFlag, resourceIDFlag, subscriptionIDFlag,
	nodeResourceGroupFlag, vmssFlag, instanceIDFlag string,
) error {
	if err := os.MkdirAll(Dir(), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	if err := c.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading config: %w", err)
	}
	// delete the node config if it already exists to avoid mutually exclusive config
	if c.IsSet("nodes." + nodeName) {
		if err := c.deleteNodeConfig(nodeName); err != nil {
			return fmt.Errorf("deleting node config: %w", err)
		}
	}

	if nodeFlag != "" {
		c.Set("nodes."+nodeName+".node", nodeFlag)
	} else if resourceIDFlag != "" {
		c.Set("nodes."+nodeName+".id", resourceIDFlag)
	} else {
		c.Set("nodes."+nodeName+".subscription", subscriptionIDFlag)
		c.Set("nodes."+nodeName+".node-resource-group", nodeResourceGroupFlag)
		c.Set("nodes."+nodeName+".vmss", vmssFlag)
		c.Set("nodes."+nodeName+".instance-id", instanceIDFlag)
	}

	if err := c.WriteConfig(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

func (c *config) deleteNodeConfig(targetNode string) error {
	return c.deleteConfig(func(settings map[string]interface{}) {
		delete(settings["nodes"].(map[string]interface{}), targetNode)
	})
}

func (c *config) deleteCurrentNodeConfig() error {
	return c.deleteConfig(func(settings map[string]interface{}) {
		delete(settings, currentNodeKey)
	})
}

// deleteConfig, viper does not support deleting a config with key, so we get
// the underlying map, delete the key via deleteKey and then re-read the config
// https://github.com/spf13/viper/issues/632#issuecomment-869668629
func (c *config) deleteConfig(deleteKey func(setting map[string]interface{})) error {
	settings := c.AllSettings()
	deleteKey(settings)
	data, err := json.MarshalIndent(settings, "", " ")
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}
	if err = c.ReadConfig(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("reading config: %w", err)
	}
	return nil
}
