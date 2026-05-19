// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package check

import (
	"fmt"
	"strings"

	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
)

func init() {
	Register(&dnsResolution{})
}

type dnsResolution struct{}

func (c *dnsResolution) Name() string { return "dns-resolution" }
func (c *dnsResolution) Description() string {
	return "Check if the node can resolve required AKS FQDNs"
}
func (c *dnsResolution) Mode() Mode { return ModeVerify }

// requiredFQDNs maps each required domain to a description of why it's needed.
var requiredFQDNs = map[string]string{
	"mcr.microsoft.com":             "Required to access images in Microsoft Container Registry (MCR)",
	"eastus.data.mcr.microsoft.com": "Required for MCR storage backed by the Azure CDN (*.data.mcr.microsoft.com)",
	"login.microsoftonline.com":     "Required for Microsoft Entra authentication",
	"packages.microsoft.com":        "Required to download packages (Moby, PowerShell, Azure CLI) via apt-get",
	"packages.aks.azure.com":        "Required to download and install required binaries (kubenet, Azure CNI)",
}

func (c *dnsResolution) Command() string {
	// Test resolution of FQDNs required for AKS node operations.
	// See https://learn.microsoft.com/en-us/azure/aks/outbound-rules-control-egress
	return `for domain in mcr.microsoft.com eastus.data.mcr.microsoft.com login.microsoftonline.com packages.microsoft.com packages.aks.azure.com; do
  if nslookup "$domain" >/dev/null 2>&1; then
    echo "$domain:ok"
  else
    echo "$domain:fail"
  fi
done`
}

func (c *dnsResolution) Parse(res *pkgruntime.RunResult) (*Result, error) {
	lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
	var failures []string
	var details []string
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		domain, status := parts[0], parts[1]
		if status == "fail" {
			failures = append(failures, domain)
			if desc, ok := requiredFQDNs[domain]; ok {
				details = append(details, fmt.Sprintf("  ✗ %s — %s", domain, desc))
			} else {
				details = append(details, fmt.Sprintf("  ✗ %s", domain))
			}
		}
	}
	// Build a list of all FQDNs tried.
	var tried []string
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			tried = append(tried, parts[0])
		}
	}
	triedStr := strings.Join(tried, ", ")

	if len(failures) > 0 {
		return &Result{
			Success: false,
			Message: fmt.Sprintf("DNS resolution failed for %d/%d required FQDN(s)", len(failures), len(requiredFQDNs)),
			Details: fmt.Sprintf("FQDNs tried: %s\n%s", triedStr, strings.Join(details, "\n")),
		}, nil
	}
	return &Result{
		Success: true,
		Message: fmt.Sprintf("DNS resolution: all %d required FQDNs resolved successfully", len(requiredFQDNs)),
		Details: fmt.Sprintf("FQDNs tried: %s", triedStr),
	}, nil
}
