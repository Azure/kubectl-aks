// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package integration

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Azure/kubectl-aks/cmd/utils"
	"github.com/Azure/kubectl-aks/cmd/utils/config"
	"github.com/stretchr/testify/require"
)

var integration = flag.Bool("integration", false, "run integration tests")

func TestMain(m *testing.M) {
	flag.Parse()
	if !*integration {
		fmt.Println("Skipping integration test.")
		os.Exit(0)
	}

	if os.Getenv("KUBECTL_AKS") == "" {
		fmt.Fprintf(os.Stderr, "KUBECTL_AKS environment variable must be set to the path of the kubectl-aks binary\n")
		os.Exit(1)
	}

	// unset existing config to avoid conflicts with node flags
	// TODO: use a different config file for integration tests
	if err := config.New().UnsetAllConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Running integration tests")
	m.Run()
}

func TestCheckAPIServerConnectivity(t *testing.T) {
	out, err := runKubectlAKS(t, "check-apiserver-connectivity")
	require.Empty(t, err, "runKubectlAKS() = %v, want nil", err)
	require.Contains(t, out, "Connectivity check: succeeded")
}

func TestRunCommandOutput(t *testing.T) {
	// test stdout
	stdout, stderr := runKubectlAKS(t, "run-command", "echo -n test")
	require.Empty(t, stderr, "runKubectlAKS() = %v, want nil", stderr)
	require.Equal(t, stdout, "test", "runKubectlAKS() = %v, want %v", stdout, "test")

	// test stderr
	stdout, stderr = runKubectlAKS(t, "run-command", "echo -n test >&2")
	require.Empty(t, stdout, "runKubectlAKS() = %v, want nil", stdout)
	require.Equal(t, stderr, "test", "runKubectlAKS() = %v, want %v", stderr, "test")
}

func TestRunCommandTimeout(t *testing.T) {
	// TODO: Investigate why this test fails on Windows after upgrading to Go 1.23
	// https://github.com/Azure/kubectl-aks/pull/69
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows due to timeout issue")
	}

	ch := make(chan struct{})
	go func() {
		runKubectlAKS(t, "run-command", "sleep inf", "--timeout", "2")
		ch <- struct{}{}
	}()
	select {
	case <-ch:
	case <-time.After(60 * time.Second):
		t.Fatal("timed out waiting for command to finish")
	}
}

func TestConfigImport(t *testing.T) {
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	resourceGroup := os.Getenv("AZURE_RESOURCE_GROUP")
	clusterName := os.Getenv("AZURE_CLUSTER_NAME")
	if subscriptionID == "" || resourceGroup == "" || clusterName == "" {
		t.Fatal("AZURE_SUBSCRIPTION_ID, AZURE_RESOURCE_GROUP, and AZURE_CLUSTER_NAME environment variables must be set to run this test")
	}

	configPath := filepath.Join(config.Dir(), "config.yaml")
	defer os.Remove(configPath)

	runCommand(t, os.Getenv("KUBECTL_AKS"), "config", "import")
	k8sConfigFile, err := os.ReadFile(configPath)
	require.Nil(t, err, "reading config file: %v", err)
	require.NotEmpty(t, k8sConfigFile, "config file is empty")

	runCommand(t, os.Getenv("KUBECTL_AKS"), "config", "unset-all")
	_, err = os.ReadFile(configPath)
	require.NotNil(t, err, "reading config file: %v", err)

	runCommand(t, os.Getenv("KUBECTL_AKS"), "config", "import",
		"--"+utils.SubscriptionIDKey, subscriptionID,
		"--"+utils.ResourceGroupKey, resourceGroup,
		"--"+utils.ClusterNameKey, clusterName,
	)
	azureConfigFile, err := os.ReadFile(configPath)
	require.Nil(t, err, "reading config file: %v", err)
	require.NotEmpty(t, azureConfigFile, "config file is empty")
	require.Equal(t, string(k8sConfigFile), string(azureConfigFile), "config file is not the same")
}
