// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
)

func TestApiserverConnectivityParse(t *testing.T) {
	c := &apiserverConnectivity{}

	tests := []struct {
		name    string
		stdout  string
		success bool
	}{
		{"success", "0", true},
		{"failure", "1", false},
		{"failure non-zero", "127", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := c.Parse(&pkgruntime.RunResult{Stdout: tt.stdout})
			require.NoError(t, err)
			assert.Equal(t, tt.success, res.Success)
		})
	}
}

func TestDNSResolutionParse(t *testing.T) {
	c := &dnsResolution{}

	t.Run("all ok", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{
			Stdout: "mcr.microsoft.com:ok\neastus.data.mcr.microsoft.com:ok\nlogin.microsoftonline.com:ok\npackages.microsoft.com:ok\npackages.aks.azure.com:ok\n",
		})
		require.NoError(t, err)
		assert.True(t, res.Success)
		assert.Contains(t, res.Message, "5 required FQDNs")
	})

	t.Run("one failure", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{
			Stdout: "mcr.microsoft.com:fail\neastus.data.mcr.microsoft.com:ok\nlogin.microsoftonline.com:ok\npackages.microsoft.com:ok\npackages.aks.azure.com:ok\n",
		})
		require.NoError(t, err)
		assert.False(t, res.Success)
		assert.Contains(t, res.Message, "1/5")
		assert.Contains(t, res.Details, "mcr.microsoft.com")
		assert.Contains(t, res.Details, "Microsoft Container Registry")
	})
}

func TestOOMEventsParse(t *testing.T) {
	c := &oomEvents{}

	t.Run("no events", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{Stdout: "EXIT:1\n"})
		require.NoError(t, err)
		assert.True(t, res.Success)
	})

	t.Run("has events", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{
			Stdout: "2024-01-01T00:00:00+0000 Out of memory: Killed process 1234\nEXIT:0\n",
		})
		require.NoError(t, err)
		assert.False(t, res.Success)
		assert.Contains(t, res.Message, "1 OOM kill event")
	})
}

func TestDiskPressureParse(t *testing.T) {
	c := &diskPressure{}

	t.Run("healthy", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{Stdout: "disk:42\ninode:5\n"})
		require.NoError(t, err)
		assert.True(t, res.Success)
		assert.Contains(t, res.Message, "42%")
	})

	t.Run("disk pressure", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{Stdout: "disk:92\ninode:5\n"})
		require.NoError(t, err)
		assert.False(t, res.Success)
		assert.Contains(t, res.Message, "92%")
	})

	t.Run("inode pressure", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{Stdout: "disk:42\ninode:90\n"})
		require.NoError(t, err)
		assert.False(t, res.Success)
		assert.Contains(t, res.Message, "inode")
	})
}

func TestProcessHealthParse(t *testing.T) {
	c := &processHealth{}

	t.Run("all active", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{Stdout: "kubelet:active\ncontainerd:active\n"})
		require.NoError(t, err)
		assert.True(t, res.Success)
	})

	t.Run("kubelet failed", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{Stdout: "kubelet:failed\ncontainerd:active\n"})
		require.NoError(t, err)
		assert.False(t, res.Success)
		assert.Contains(t, res.Message, "kubelet")
	})
}

func TestDNSTraceParse(t *testing.T) {
	c := &failedDNSTrace{}

	t.Run("no failures", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{Stdout: ""})
		require.NoError(t, err)
		assert.True(t, res.Success)
	})

	t.Run("dns failures", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{
			Stdout: `{"name":"bad.example.com.","rcode":"NameError","qtype":"A","nameserver":{"addr":"168.63.129.16","version":4},"proc":{"comm":"curl","pid":1234},"latency_ns":"40.48ms","src":{"addr":"168.63.129.16","port":53},"dst":{"addr":"10.0.0.5","port":36026}}`,
		})
		require.NoError(t, err)
		assert.False(t, res.Success)
		assert.Contains(t, res.Message, "1 DNS failure")
		assert.Contains(t, res.Details, "bad.example.com.")
		assert.Contains(t, res.Details, "NameError")
		assert.Contains(t, res.Details, "168.63.129.16")
		assert.Contains(t, res.Details, "curl(1234)")
		assert.NotContains(t, res.Details, "LATENCY")
	})
}

func TestDNSSlowTraceParse(t *testing.T) {
	c := &dnsSlowTrace{}

	t.Run("no slow queries", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{Stdout: ""})
		require.NoError(t, err)
		assert.True(t, res.Success)
	})

	t.Run("slow queries detected", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{
			Stdout: `{"name":"slow.example.com.","rcode":"Success","qtype":"A","nameserver":{"addr":"168.63.129.16","version":4},"proc":{"comm":"curl","pid":5678},"latency_ns":"1.2s","src":{"addr":"168.63.129.16","port":53},"dst":{"addr":"10.0.0.5","port":36026}}`,
		})
		require.NoError(t, err)
		assert.False(t, res.Success)
		assert.Contains(t, res.Message, "1 slow DNS")
		assert.Contains(t, res.Details, "slow.example.com.")
		assert.Contains(t, res.Details, "168.63.129.16")
		assert.Contains(t, res.Details, "curl(5678)")
		assert.Contains(t, res.Details, "1.2s")
		assert.Contains(t, res.Details, "LATENCY")
	})
}

func TestTCPDropTraceParse(t *testing.T) {
	c := &tcpDropTrace{}

	t.Run("no events", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{Stdout: ""})
		require.NoError(t, err)
		assert.True(t, res.Success)
	})

	t.Run("losses detected", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{
			Stdout: `{"src":{"addr":"10.244.0.160","port":59344},"dst":{"addr":"20.105.36.95","port":443},"proc":{"comm":"proxy-agent","pid":33571},"state":1,"reason":0,"type":"LOSS","tcpflags":"","timestamp":"2026-05-15T07:59:43.603Z"}`,
		})
		require.NoError(t, err)
		assert.False(t, res.Success)
		assert.Contains(t, res.Message, "1 packet loss")
		assert.Contains(t, res.Details, "proxy-agent(33571)")
		assert.Contains(t, res.Details, "10.244.0.160:59344")
		assert.Contains(t, res.Details, "20.105.36.95:443")
	})
}

func TestTCPRetransTraceParse(t *testing.T) {
	c := &tcpRetransTrace{}

	t.Run("no events", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{Stdout: ""})
		require.NoError(t, err)
		assert.True(t, res.Success)
	})

	t.Run("retransmissions detected", func(t *testing.T) {
		res, err := c.Parse(&pkgruntime.RunResult{
			Stdout: `{"src":{"addr":"10.244.0.160","port":59344},"dst":{"addr":"20.105.36.95","port":443},"proc":{"comm":"proxy-agent","pid":33571},"state":1,"reason":0,"type":"RETRANS","tcpflags":"PSH|ACK","timestamp":"2026-05-15T07:59:43.603Z"}`,
		})
		require.NoError(t, err)
		assert.False(t, res.Success)
		assert.Contains(t, res.Message, "1 retransmission")
		assert.Contains(t, res.Details, "proxy-agent(33571)")
		assert.Contains(t, res.Details, "PSH|ACK")
		assert.Contains(t, res.Details, "20.105.36.95:443")
	})
}

func TestRegistry(t *testing.T) {
	all := All()
	assert.NotEmpty(t, all)

	verifyChecks := ByMode(ModeVerify)
	traceChecks := ByMode(ModeTrace)
	assert.Equal(t, len(all), len(verifyChecks)+len(traceChecks))

	// Verify expected checks exist
	names := make(map[string]bool)
	for _, c := range all {
		names[c.Name()] = true
	}
	assert.True(t, names["apiserver-connectivity"])
	assert.True(t, names["dns-resolution"])
	assert.True(t, names["oom-events"])
	assert.True(t, names["disk-pressure"])
	assert.True(t, names["process-health"])
	assert.True(t, names["dns-failed"])
	assert.True(t, names["dns-slow"])
	assert.True(t, names["tcp-drops"])
	assert.True(t, names["tcp-retrans"])
}

func TestFormatResults(t *testing.T) {
	t.Run("multi node has separators", func(t *testing.T) {
		results := []NodeResult{
			{NodeName: "node1", Result: &Result{Success: true, Message: "ok"}},
			{NodeName: "node2", Result: &Result{Success: false, Message: "fail", Details: "some detail"}},
		}
		output, hasFailure := FormatResults(results)
		assert.True(t, hasFailure)
		assert.Contains(t, output, "=== node1 ===")
		assert.Contains(t, output, "=== node2 ===")
		assert.Contains(t, output, "✓")
		assert.Contains(t, output, "✗")
		assert.Contains(t, output, "some detail")
	})

	t.Run("single node no separator", func(t *testing.T) {
		results := []NodeResult{
			{NodeName: "node1", Result: &Result{Success: true, Message: "ok"}},
		}
		output, _ := FormatResults(results)
		assert.NotContains(t, output, "===")
		assert.Contains(t, output, "✓ ok")
	})
}
