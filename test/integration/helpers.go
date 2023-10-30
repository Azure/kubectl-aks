// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package integration

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/kinvolk/inspektor-gadget/pkg/k8sutil"
	"github.com/stretchr/testify/require"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func runKubectlAKS(t *testing.T, args ...string) (string, string) {
	t.Helper()

	args = append(nodeFlag(t), args...)
	return runCommand(t, os.Getenv("KUBECTL_AKS"), args...)
}

func runCommand(t *testing.T, name string, args ...string) (string, string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	t.Logf("Running command: %s", cmd.String())
	err := cmd.Run()
	t.Logf("Command output: \n%s", stdout.String())
	t.Logf("Command error: \n%s", stderr.String())
	require.Nil(t, err, "cmd.Run() = %v, want nil", err)

	return stdout.String(), stderr.String()
}

func nodeFlag(t *testing.T) []string {
	t.Helper()

	if envNodeName := os.Getenv("AZURE_NODE_NAME"); envNodeName != "" {
		return []string{"--node", envNodeName}
	}

	clientset, err := k8sutil.NewClientsetFromConfigFlags(genericclioptions.NewConfigFlags(false))
	require.Nil(t, err, "k8sutil.NewClientsetFromConfigFlags() = %v, want nil", err)

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metaV1.ListOptions{})
	require.Nil(t, err, "clientset.CoreV1().Nodes().List() = %v, want nil", err)
	require.NotEmpty(t, nodes.Items, "nodes.Items = %v, want not empty", nodes.Items)

	return []string{"--node", nodes.Items[0].Name}
}
