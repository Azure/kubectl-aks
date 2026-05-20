// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

// Shared types for DNS trace checks.

type dnsEvent struct {
	Name       string      `json:"name"`
	Rcode      string      `json:"rcode"`
	QType      string      `json:"qtype"`
	Nameserver dnsAddr     `json:"nameserver"`
	Proc       dnsProc     `json:"proc"`
	K8s        K8s         `json:"k8s"`
	Src        dnsEndpoint `json:"src"`
	Dst        dnsEndpoint `json:"dst"`
	Timestamp  string      `json:"timestamp"`
	Latency    string      `json:"latency_ns"`
}

type dnsAddr struct {
	Addr string `json:"addr"`
}

type dnsEndpoint struct {
	Addr  string `json:"addr"`
	Port  int    `json:"port"`
	Proto string `json:"proto"`
}

type dnsProc struct {
	Comm string `json:"comm"`
	PID  int    `json:"pid"`
}
