// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

import (
	"fmt"
	"strings"
	"text/tabwriter"

	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
)

func init() {
	Register(newTCPDropTrace())
}

type tcpDropTrace struct {
	IGCheck
}

func newTCPDropTrace() *tcpDropTrace {
	return &tcpDropTrace{
		IGCheck: IGCheck{
			GadgetImage: "trace_tcpretrans",
			Filters:     []string{"type=LOSS"},
		},
	}
}

func (c *tcpDropTrace) Name() string { return "tcp-drops" }
func (c *tcpDropTrace) Description() string {
	return "Trace TCP packet drops (losses) on the node"
}
func (c *tcpDropTrace) Mode() Mode { return ModeTrace }

func (c *tcpDropTrace) Command() string {
	return c.IGCommand()
}

func (c *tcpDropTrace) Parse(res *pkgruntime.RunResult) (*Result, error) {
	events := parseTCPEvents(res.Stdout)
	if len(events) == 0 {
		return &Result{
			Success: true,
			Message: "TCP drops trace: no packet losses observed during trace period",
		}, nil
	}

	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "SRC\tDST\tPOD\tPROCESS")
	for _, ev := range events {
		src := fmt.Sprintf("%s:%d", ev.Src.Addr, ev.Src.Port)
		dst := fmt.Sprintf("%s:%d", ev.Dst.Addr, ev.Dst.Port)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", src, dst, ev.K8s.FormatPod(), formatTCPProc(ev))
	}
	w.Flush()

	return &Result{
		Success: false,
		Message: fmt.Sprintf("TCP drops trace: %d packet loss(es) detected", len(events)),
		Details: b.String(),
	}, nil
}
