// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package config

import (
	"os"
	"path"
	"sort"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestClusterConfig(t *testing.T) {
	t.Run("UseClusterConfig", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		err := cfg.UseClusterConfig("non-existent")
		require.Error(t, err)

		err = cfg.UseClusterConfig("another-cluster")
		require.NoError(t, err)

		current := cfg.GetString(currentClusterKey)
		require.Equal(t, "another-cluster", current)
	})

	t.Run("UnsetClusterConfig", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		require.True(t, cfg.IsSet("clusters.test-cluster"))

		err := cfg.UnsetClusterConfig("test-cluster")
		require.NoError(t, err)

		require.False(t, cfg.IsSet("clusters.test-cluster"))
		// Should also unset current-cluster since it was pointing to test-cluster
		require.Equal(t, "", cfg.GetString(currentClusterKey))
	})

	t.Run("ListClusters", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		clusters, err := cfg.ListClusters()
		require.NoError(t, err)
		sort.Strings(clusters)
		require.Equal(t, []string{"another-cluster", "test-cluster"}, clusters)
	})

	t.Run("CurrentClusterName", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		name := cfg.CurrentClusterName()
		require.Equal(t, "test-cluster", name)
	})

	t.Run("ListClusterNodes", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		nodes, err := cfg.ListClusterNodes("test-cluster")
		require.NoError(t, err)
		sort.Strings(nodes)
		require.Equal(t, []string{"test-node", "test-node-2"}, nodes)
	})

	t.Run("GetClusterNodeConfig", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		nc, ok := cfg.GetClusterNodeConfig("test-cluster", "test-node")
		require.True(t, ok)
		require.Equal(t, "mySubID", nc.GetString("subscription"))
		require.Equal(t, "myRG", nc.GetString("node-resource-group"))
		require.Equal(t, "myVMSS", nc.GetString("vmss"))
		require.Equal(t, "myInsID", nc.GetString("instance-id"))

		_, ok = cfg.GetClusterNodeConfig("test-cluster", "non-existent")
		require.False(t, ok)
	})

	t.Run("SetClusterNodeConfigWithVMSSInfo", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		err := cfg.SetClusterNodeConfigWithVMSSInfo("test-cluster", "new-node", "sub1", "rg1", "vmss1", "id1")
		require.NoError(t, err)

		nc, ok := cfg.GetClusterNodeConfig("test-cluster", "new-node")
		require.True(t, ok)
		require.Equal(t, "sub1", nc.GetString("subscription"))
		require.Equal(t, "vmss1", nc.GetString("vmss"))
	})

	t.Run("DeleteClusterNode", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		err := cfg.DeleteClusterNode("test-cluster", "test-node-2")
		require.NoError(t, err)

		_, ok := cfg.GetClusterNodeConfig("test-cluster", "test-node-2")
		require.False(t, ok)

		// Other node should still exist
		_, ok = cfg.GetClusterNodeConfig("test-cluster", "test-node")
		require.True(t, ok)
	})

	t.Run("SetClusterMetadata", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		err := cfg.SetClusterMetadata("brand-new", "newSub", "newRG")
		require.NoError(t, err)

		clusters, err := cfg.ListClusters()
		require.NoError(t, err)
		require.Contains(t, clusters, "brand-new")
	})

	t.Run("UnsetCurrentClusterConfig", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		require.Equal(t, "test-cluster", cfg.GetString(currentClusterKey))

		err := cfg.UnsetCurrentClusterConfig()
		require.NoError(t, err)

		require.Equal(t, "", cfg.GetString(currentClusterKey))
	})
}

func TestIsLegacyConfig(t *testing.T) {
	t.Run("OldFormatIsLegacy", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadTempConfig(t) // uses old-style testdata/config.yaml
		require.True(t, cfg.IsLegacyConfig())
	})

	t.Run("NewFormatIsNotLegacy", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)
		require.False(t, cfg.IsLegacyConfig())
	})
}

func TestClusterAwareNodeLookup(t *testing.T) {
	t.Run("GetNodeConfig_FindsInCluster", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		nc, ok := cfg.GetNodeConfig("test-node")
		require.True(t, ok)
		require.Equal(t, "myVMSS", nc.GetString("vmss"))
	})

	t.Run("CurrentConfig_ResolvesThroughCluster", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		cc, ok := cfg.CurrentConfig()
		require.True(t, ok)
		require.Equal(t, "myVMSS", cc.GetString("vmss"))
	})

	t.Run("UseNodeConfig_FindsInCluster", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		err := cfg.UseNodeConfig("other-node")
		require.NoError(t, err)

		require.Equal(t, "other-node", cfg.GetString(currentNodeKey))
	})

	t.Run("UseNodeConfig_NotFound", func(t *testing.T) {
		t.Parallel()
		cfg := createAndReadClusterConfig(t)

		err := cfg.UseNodeConfig("ghost-node")
		require.Error(t, err)
	})
}

func createAndReadClusterConfig(t *testing.T) *Config {
	tempDir := t.TempDir()
	cfgPath := path.Join(tempDir, configFilename)
	v := viper.New()
	v.SetConfigFile(cfgPath)
	cfg := &Config{
		Viper:      v,
		configPath: cfgPath,
	}

	data, err := os.ReadFile(path.Join("testdata", "cluster_config.yaml"))
	require.NoError(t, err)
	require.NotEmpty(t, data)

	err = os.WriteFile(cfgPath, data, 0o600)
	require.NoError(t, err)

	err = cfg.ReadInConfig()
	require.NoError(t, err)

	return cfg
}
