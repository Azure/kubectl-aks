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
	Register(newFailedDNSTrace())
}

type failedDNSTrace struct {
	IGCheck
}

func newFailedDNSTrace() *failedDNSTrace {
	return &failedDNSTrace{
		IGCheck: IGCheck{
			GadgetImage: "trace_dns",
			Filters:     []string{"rcode!=Success", "qr=R"},
		},
	}
}

func (c *failedDNSTrace) Name() string { return "dns-failed" }
func (c *failedDNSTrace) Description() string {
	return "Trace DNS queries on the node for a given duration, reporting failures"
}
func (c *failedDNSTrace) Mode() Mode { return ModeTrace }

func (c *failedDNSTrace) Command() string {
	return c.IGCommand()
}

func (c *failedDNSTrace) Parse(res *pkgruntime.RunResult) (*Result, error) {
	stdout := strings.TrimSpace(res.Stdout)
	if stdout == "" {
		return &Result{
			Success: true,
			Message: "DNS trace: no DNS failures observed during trace period",
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
			Message: "DNS trace: no DNS failures observed during trace period",
		}, nil
	}

	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tRCODE\tQTYPE\tNAMESERVER\tPOD\tPROCESS")
	for _, ev := range events {
		proc := ev.Proc.Comm
		if ev.Proc.PID > 0 {
			proc = fmt.Sprintf("%s(%d)", ev.Proc.Comm, ev.Proc.PID)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			ev.Name, ev.Rcode, ev.QType, ev.Nameserver.Addr, ev.K8s.FormatPod(), proc)
	}
	w.Flush()

	return &Result{
		Success: false,
		Message: fmt.Sprintf("DNS trace: %d DNS failure(s) observed", len(events)),
		Details: b.String(),
	}, nil
}
