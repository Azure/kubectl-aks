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
	Register(newTCPRetransTrace())
}

type tcpRetransTrace struct {
	IGCheck
}

func newTCPRetransTrace() *tcpRetransTrace {
	return &tcpRetransTrace{
		IGCheck: IGCheck{
			GadgetImage: "trace_tcpretrans",
			Filters:     []string{"type=RETRANS"},
		},
	}
}

func (c *tcpRetransTrace) Name() string { return "tcp-retrans" }
func (c *tcpRetransTrace) Description() string {
	return "Trace TCP retransmissions on the node"
}
func (c *tcpRetransTrace) Mode() Mode { return ModeTrace }

func (c *tcpRetransTrace) Command() string {
	return c.IGCommand()
}

func (c *tcpRetransTrace) Parse(res *pkgruntime.RunResult) (*Result, error) {
	events := parseTCPEvents(res.Stdout)
	if len(events) == 0 {
		return &Result{
			Success: true,
			Message: "TCP retrans trace: no retransmissions observed during trace period",
		}, nil
	}

	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "SRC\tDST\tFLAGS\tPOD\tPROCESS")
	for _, ev := range events {
		src := fmt.Sprintf("%s:%d", ev.Src.Addr, ev.Src.Port)
		dst := fmt.Sprintf("%s:%d", ev.Dst.Addr, ev.Dst.Port)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", src, dst, ev.TCPFlags, ev.K8s.FormatPod(), formatTCPProc(ev))
	}
	w.Flush()

	return &Result{
		Success: false,
		Message: fmt.Sprintf("TCP retrans trace: %d retransmission(s) detected", len(events)),
		Details: b.String(),
	}, nil
}
