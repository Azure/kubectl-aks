// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package integration

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var integration = flag.Bool("integration", false, "run integration tests")

func TestMain(m *testing.M) {
	flag.Parse()
	if !*integration {
		fmt.Println("Skipping integration test.")
		os.Exit(0)
	}

	if os.Getenv("KUBECTL_AZ") == "" {
		fmt.Fprintf(os.Stderr, "KUBECTL_AZ environment variable must be set to the path of the kubectl-az binary\n")
		os.Exit(1)
	}

	fmt.Println("Running integration tests")
	m.Run()
}

func TestCheckAPIServerConnectivity(t *testing.T) {
	out := runKubectlAZ(t, "check-apiserver-connectivity")
	require.Contains(t, out, "Connectivity check: succeeded")
}

func TestRunCommand(t *testing.T) {
	// test stdout
	out := runKubectlAZ(t, "run-command", "echo test")
	stdout, stderr := parseRunCommand(t, out)
	require.Equal(t, stdout, "test", "parseRunCommand() = %v, want %v", stdout, "test")
	require.Empty(t, stderr, "parseRunCommand() = %v, want %v", stderr, "")

	// test stderr
	out = runKubectlAZ(t, "run-command", "echo test >&2")
	stdout, stderr = parseRunCommand(t, out)
	require.Empty(t, stdout, "parseRunCommand() = %v, want %v", stdout, "")
	require.Equal(t, stderr, "test", "parseRunCommand() = %v, want %v", stderr, "test")
}

func parseRunCommand(t *testing.T, out string) (string, string) {
	split := regexp.MustCompile(`(\[(stdout|stderr)\])`).Split(out, -1)
	require.Len(t, split, 3, "couldn't parse response message:\n%s", out)
	stdOutput := strings.TrimSpace(split[1])
	stdError := strings.TrimSpace(split[2])
	return stdOutput, stdError
}
