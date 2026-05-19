// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

import (
	"fmt"
	"strings"
)

// IGCheck provides common command-building logic for checks that use the ig binary.
// Embed this in a check struct to automatically generate the ig command string.
type IGCheck struct {
	// GadgetImage is the gadget to run (e.g. "trace_dns", "trace_tcpretrans").
	GadgetImage string
	// OutputMode is the output format flag (default: "json").
	OutputMode string
	// Filters is a list of filter expressions joined by comma for --filter.
	// It is recommended to use the filter to limit the data volume since
	// azure-api runtime only has 4KB of output buffer.
	Filters []string
	// ExtraArgs are additional arguments appended after the filter flags.
	ExtraArgs []string
}

// IGCommand builds the ig command string with the standard flags.
// The returned command uses {{.Duration}} as a placeholder for the trace timeout.
func (ig *IGCheck) IGCommand() string {
	outputMode := ig.OutputMode
	if outputMode == "" {
		outputMode = "json"
	}

	var b strings.Builder
	b.WriteString("ig run ")
	b.WriteString(ig.GadgetImage)
	b.WriteString(" --host --timeout {{.Duration}} --output ")
	b.WriteString(outputMode)

	if len(ig.Filters) > 0 {
		b.WriteString(" --filter ")
		b.WriteString(strings.Join(ig.Filters, ","))
	}

	for _, arg := range ig.ExtraArgs {
		b.WriteString(" ")
		b.WriteString(arg)
	}

	b.WriteString(" 2>/dev/null || true")
	return b.String()
}

// K8s holds Kubernetes metadata from ig event output.
type K8s struct {
	Namespace string `json:"namespace"`
	PodName   string `json:"podName"`
}

// FormatPod returns "namespace/pod" or empty string if no k8s context.
func (k K8s) FormatPod() string {
	if k.Namespace == "" && k.PodName == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", k.Namespace, k.PodName)
}
