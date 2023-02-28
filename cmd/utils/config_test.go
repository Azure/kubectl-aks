// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"os"
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("TestUseNodeConfig", func(t *testing.T) {
		config = viper.New()
		config.SetConfigFile(path.Join("testdata", configFilename))

		err := UseNodeConfig("not-found")
		require.NotNil(t, err, "UseNodeConfig() = nil, want error")

		err = UseNodeConfig("test-node")
		require.Nil(t, err, "UseNodeConfig() = %v, want nil", err)

		expectedNode := "test-node"
		currentNode := config.GetString(currentNodeKey)
		require.Equal(t, expectedNode, currentNode, "config.GetString(currentNodeKey) = %v, want %v", currentNode, expectedNode)
	})
	t.Run("TestUnsetCurrentNodeConfig", func(t *testing.T) {
		err := createAndLoadTempConfig(t)
		require.Nil(t, err, "createAndLoadTempConfig() = %v, want nil", err)

		currentNode := config.IsSet(currentNodeKey)
		require.True(t, currentNode, "config.IsSet(currentNodeKey) = %v, want true", currentNode)

		err = UnsetCurrentNodeConfig()
		require.Nil(t, err, "UnsetCurrentNodeConfig() = %v, want nil", err)

		currentNode = config.IsSet(currentNodeKey)
		require.False(t, currentNode, "config.IsSet(currentNodeKey) = %v, want false", currentNode)
	})
	t.Run("TestUnsetNodeConfig", func(t *testing.T) {
		err := createAndLoadTempConfig(t)
		require.Nil(t, err, "createAndLoadTempConfig() = %v, want nil", err)

		testNode := config.IsSet("nodes.test-node")
		require.True(t, testNode, "config.IsSet(nodes.test-node) = %v, want true", testNode)

		err = UnsetNodeConfig("test-node")
		require.Nil(t, err, "UnsetNodeConfig() = %v, want nil", err)

		testNode = config.IsSet("nodes.test-node")
		require.False(t, testNode, "config.IsSet(nodes.test-node) = %v, want false", testNode)
	})
}

func createAndLoadTempConfig(t *testing.T) error {
	tempDir := t.TempDir()
	config = viper.New()
	config.SetConfigFile(path.Join(tempDir, configFilename))
	cfg, err := os.ReadFile(path.Join("testdata", configFilename))
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}
	if err = os.WriteFile(path.Join(tempDir, configFilename), cfg, 0600); err != nil {
		t.Fatalf("writing config file: %v", err)
	}
	if err = loadConfig(); err != nil {
		t.Fatalf("loadConfig() = %v", err)
	}
	return err
}
