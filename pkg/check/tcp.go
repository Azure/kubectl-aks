// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

import (
	"encoding/json"
	"fmt"
	"strings"
)

type tcpEndpoint struct {
	Addr  string `json:"addr"`
	Port  int    `json:"port"`
	Proto string `json:"proto"`
}

type tcpProc struct {
	Comm string `json:"comm"`
	PID  int    `json:"pid"`
}

type tcpEvent struct {
	Src       tcpEndpoint `json:"src"`
	Dst       tcpEndpoint `json:"dst"`
	Proc      tcpProc     `json:"proc"`
	K8s       K8s         `json:"k8s"`
	State     int         `json:"state"`
	Reason    int         `json:"reason"`
	Type      string      `json:"type"`
	TCPFlags  string      `json:"tcpflags"`
	Timestamp string      `json:"timestamp"`
}

func formatTCPProc(ev tcpEvent) string {
	if ev.Proc.PID > 0 {
		return fmt.Sprintf("%s(%d)", ev.Proc.Comm, ev.Proc.PID)
	}
	return ev.Proc.Comm
}

// parseTCPEvents parses newline-delimited JSON output from ig trace_tcpretrans.
func parseTCPEvents(stdout string) []tcpEvent {
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return nil
	}
	var events []tcpEvent
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ev tcpEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		events = append(events, ev)
	}
	return events
}
