// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package config

import (
	"os"
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestNodeConfig(t *testing.T) {
	t.Run("TestUseNodeConfig", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadTempConfig(t)
		require.NotNil(t, cfg, "createAndReadTempConfig() = %v, want not nil", cfg)

		err := cfg.UseNodeConfig("not-found")
		require.NotNil(t, err, "UseNodeConfig() = nil, want error")

		err = cfg.UseNodeConfig("test-node")
		require.Nil(t, err, "UseNodeConfig() = %v, want nil", err)

		expectedNode := "test-node"
		currentNode := cfg.GetString(currentNodeKey)
		require.Equal(t, expectedNode, currentNode, "cfg.GetString(currentNodeKey) = %v, want %v", currentNode, expectedNode)
	})
	t.Run("TestUnsetCurrentNodeConfig", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadTempConfig(t)
		require.NotNil(t, cfg, "createAndReadTempConfig() = %v, want not nil", cfg)

		currentNode := cfg.IsSet(currentNodeKey)
		require.True(t, currentNode, "cfg.IsSet(currentNodeKey) = %v, want true", currentNode)

		err := cfg.UnsetCurrentNodeConfig()
		require.Nil(t, err, "cfg.UnsetCurrentNodeConfig() = %v, want nil", err)

		currentNode = cfg.IsSet(currentNodeKey)
		require.False(t, currentNode, "cfg.IsSet(currentNodeKey) = %v, want false", currentNode)
	})
	t.Run("TestUnsetNodeConfig", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadTempConfig(t)
		require.NotNil(t, cfg, "createAndReadTempConfig() = %v, want not nil", cfg)

		testNode := cfg.IsSet("nodes.test-node")
		require.True(t, testNode, "cfg.IsSet(nodes.test-node) = %v, want true", testNode)

		err := cfg.UnsetNodeConfig("test-node")
		require.Nil(t, err, "cfg.UnsetNodeConfig() = %v, want nil", err)

		testNode = cfg.IsSet("nodes.test-node")
		require.False(t, testNode, "cfg.IsSet(nodes.test-node) = %v, want false", testNode)
	})
	t.Run("SetNodeConfigWithNodeFlag", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadTempConfig(t)
		require.NotNil(t, cfg, "createAndReadTempConfig() = %v, want not nil", cfg)

		expectedNode := "test-new-node"
		err := cfg.SetNodeConfigWithNodeFlag("test-new-config", expectedNode)
		require.Nil(t, err, "cfg.SetNodeConfigWithNodeFlag() = %v, want nil", err)

		testNode := cfg.IsSet("nodes.test-new-config")
		require.True(t, testNode, "cfg.IsSet(nodes.test-new-config) = %v, want true", testNode)

		testNodeName := cfg.GetString("nodes.test-new-config.node")
		require.Equal(t, "test-new-node", testNodeName, "cfg.GetString(nodes.test-new-config.name) = %v, want %v", testNodeName, expectedNode)
	})
	t.Run("SetNodeConfigWithResourceIDFlag", func(t *testing.T) {
		t.Parallel()

		cfg := createAndReadTempConfig(t)
		require.NotNil(t, cfg, "createAndReadTempConfig() = %v, want not nil", cfg)

		expectedResourceID := "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ContainerService/managedClusters/test-cluster"
		err := cfg.SetNodeConfigWithResourceIDFlag("test-new-config", expectedResourceID)
		require.Nil(t, err, "cfg.SetNodeConfigWithResourceIDFlag() = %v, want nil", err)

		testNode := cfg.IsSet("nodes.test-new-config")
		require.True(t, testNode, "cfg.IsSet(nodes.test-new-config) = %v, want true", testNode)

		testResourceID := cfg.GetString("nodes.test-new-config.id")
		require.Equal(t, expectedResourceID, testResourceID, "cfg.GetString(nodes.test-new-config.name) = %v, want %v", testResourceID, expectedResourceID)
	})
	t.Run("SetNodeConfigWithVMSSInfoFlag", func(t *testing.T) {
		t.Parallel()

		cfg := createAndReadTempConfig(t)
		require.NotNil(t, cfg, "createAndReadTempConfig() = %v, want not nil", cfg)

		expectedSubscriptionID := "00000000-0000-0000-0000-000000000000"
		expectedNodeResourceGroup := "test-rg"
		expectedVMSS := "test-vmss"
		expectedInstanceID := "0"
		err := cfg.SetNodeConfigWithVMSSInfoFlag("test-new-node", expectedSubscriptionID, expectedNodeResourceGroup, expectedVMSS, expectedInstanceID)
		require.Nil(t, err, "cfg.SetNodeConfigWithVMSSInfoFlag() = %v, want nil", err)

		testNode := cfg.IsSet("nodes.test-new-node")
		require.True(t, testNode, "cfg.IsSet(nodes.test-new-node) = %v, want true", testNode)

		testSubscriptionID := cfg.GetString("nodes.test-new-node.subscription")
		require.Equal(t, expectedSubscriptionID, testSubscriptionID, "cfg.GetString(nodes.test-new-node.subscription-id) = %v, want %v", testSubscriptionID, expectedSubscriptionID)

		testNodeResourceGroup := cfg.GetString("nodes.test-new-node.node-resource-group")
		require.Equal(t, expectedNodeResourceGroup, testNodeResourceGroup, "cfg.GetString(nodes.test-new-node.node-resource-group) = %v, want %v", testNodeResourceGroup, expectedNodeResourceGroup)

		testVMSS := cfg.GetString("nodes.test-new-node.vmss")
		require.Equal(t, expectedVMSS, testVMSS, "cfg.GetString(nodes.test-new-node.vmss) = %v, want %v", testVMSS, expectedVMSS)

		testInstanceID := cfg.GetString("nodes.test-new-node.instance-id")
		require.Equal(t, expectedInstanceID, testInstanceID, "cfg.GetString(nodes.test-new-node.instance-id) = %v, want %v", testInstanceID, expectedInstanceID)
	})
}

func createAndReadTempConfig(t *testing.T) *config {
	tempDir := t.TempDir()
	cfgPath := path.Join(tempDir, configFilename)
	v := viper.New()
	v.SetConfigFile(cfgPath)
	cfg := &config{
		Viper:      v,
		configPath: cfgPath,
	}

	data, err := os.ReadFile(path.Join("testdata", "config.yaml"))
	require.Nil(t, err, "os.ReadFile(cfgPath) err = %v, want nil", err)
	require.NotNil(t, data, "os.ReadFile(cfgPath) data = %v, want not nil", data)

	err = os.WriteFile(cfgPath, data, 0600)
	require.Nil(t, err, "os.WriteFile(cfgPath) = %v, want nil", err)

	err = cfg.ReadInConfig()
	require.Nil(t, err, "cfg.ReadInConfig() = %v, want nil", cfg)

	return cfg
}
