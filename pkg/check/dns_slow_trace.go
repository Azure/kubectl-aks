// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
)

func init() {
	Register(newDNSSlowTrace())
}

type dnsSlowTrace struct {
	IGCheck
}

func newDNSSlowTrace() *dnsSlowTrace {
	return &dnsSlowTrace{
		IGCheck: IGCheck{
			GadgetImage: "trace_dns",
			// 500ms = 500_000_000 ns. Filter for responses (qr=R) with latency >= 500ms.
			Filters: []string{"latency_ns_raw>=500000000", "qr=R"},
		},
	}
}

func (c *dnsSlowTrace) Name() string { return "dns-slow" }
func (c *dnsSlowTrace) Description() string {
	return "Trace DNS queries taking longer than 500ms"
}
func (c *dnsSlowTrace) Mode() Mode { return ModeTrace }

func (c *dnsSlowTrace) Command() string {
	return c.IGCommand()
}

func (c *dnsSlowTrace) Parse(res *pkgruntime.RunResult) (*Result, error) {
	stdout := strings.TrimSpace(res.Stdout)
	if stdout == "" {
		return &Result{
			Success: true,
			Message: "DNS slow trace: no slow DNS queries observed during trace period",
		}, nil
	}

	var events []dnsEvent
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ev dnsEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		events = append(events, ev)
	}

	if len(events) == 0 {
		return &Result{
			Success: true,
			Message: "DNS slow trace: no slow DNS queries observed during trace period",
		}, nil
	}

	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tRCODE\tQTYPE\tNAMESERVER\tLATENCY\tPOD\tPROCESS")
	for _, ev := range events {
		proc := ev.Proc.Comm
		if ev.Proc.PID > 0 {
			proc = fmt.Sprintf("%s(%d)", ev.Proc.Comm, ev.Proc.PID)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			ev.Name, ev.Rcode, ev.QType, ev.Nameserver.Addr, ev.Latency, ev.K8s.FormatPod(), proc)
	}
	w.Flush()

	return &Result{
		Success: false,
		Message: fmt.Sprintf("DNS slow trace: %d slow DNS query/queries observed (>500ms)", len(events)),
		Details: b.String(),
	}, nil
}
