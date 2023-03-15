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

func runKubectlAZ(t *testing.T, args ...string) string {
	t.Helper()

	cmd := exec.Command(os.Getenv("KUBECTL_AZ"), append(nodeFlag(t), args...)...)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	t.Logf("Running command: %s", cmd.String())
	err := cmd.Run()
	require.Empty(t, stderr.String(), "stderr.String() = %v, want empty", stderr.String())
	require.Nil(t, err, "cmd.Run() = %v, want nil", err)
	t.Logf("Command output: \n%s", stdout.String())

	return stdout.String()
}

func nodeFlag(t *testing.T) []string {
	t.Helper()

	clientset, err := k8sutil.NewClientsetFromConfigFlags(genericclioptions.NewConfigFlags(false))
	require.Nil(t, err, "k8sutil.NewClientsetFromConfigFlags() = %v, want nil", err)

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metaV1.ListOptions{})
	require.Nil(t, err, "clientset.CoreV1().Nodes().List() = %v, want nil", err)
	require.NotEmpty(t, nodes.Items, "nodes.Items = %v, want not empty", nodes.Items)

	return []string{"--node", nodes.Items[0].Name}
}
