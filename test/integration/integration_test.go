// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package integration

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

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

	fmt.Println("Running integration tests")
	m.Run()
}

func TestCheckAPIServerConnectivity(t *testing.T) {
	out := runKubectlAKS(t, "check-apiserver-connectivity")
	require.Contains(t, out, "Connectivity check: succeeded")
}

func TestRunCommandOutput(t *testing.T) {
	// test stdout
	out := runKubectlAKS(t, "run-command", "echo test")
	stdout, stderr := parseRunCommand(t, out)
	require.Equal(t, stdout, "test", "parseRunCommand() = %v, want %v", stdout, "test")
	require.Empty(t, stderr, "parseRunCommand() = %v, want %v", stderr, "")

	// test stderr
	out = runKubectlAKS(t, "run-command", "echo test >&2")
	stdout, stderr = parseRunCommand(t, out)
	require.Empty(t, stdout, "parseRunCommand() = %v, want %v", stdout, "")
	require.Equal(t, stderr, "test", "parseRunCommand() = %v, want %v", stderr, "test")
}

func TestRunCommandTimeout(t *testing.T) {
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

func parseRunCommand(t *testing.T, out string) (string, string) {
	split := regexp.MustCompile(`(\[(stdout|stderr)\])`).Split(out, -1)
	require.Len(t, split, 3, "couldn't parse response message:\n%s", out)
	stdOutput := strings.TrimSpace(split[1])
	stdError := strings.TrimSpace(split[2])
	return stdOutput, stdError
}

func TestConfigImport(t *testing.T) {
	subcriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	resourceGroup := os.Getenv("AZURE_RESOURCE_GROUP")
	clusterName := os.Getenv("AZURE_CLUSTER_NAME")
	if subcriptionID == "" || resourceGroup == "" || clusterName == "" {
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

	runCommand(t, os.Getenv("KUBECTL_AKS"), "config", "import", "-s", subcriptionID, "-g", resourceGroup, "-c", clusterName)
	azureConfigFile, err := os.ReadFile(configPath)
	require.Nil(t, err, "reading config file: %v", err)
	require.NotEmpty(t, azureConfigFile, "config file is empty")
	require.Equal(t, k8sConfigFile, azureConfigFile, "config file is not the same")
}
