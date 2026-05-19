# Check

The `check` command provides built-in diagnostic checks for AKS nodes. Checks
are organized into two categories:

- **verify** — Point-in-time checks that run a command and report pass/fail.
- **trace** — Duration-based checks that observe node activity for a specified
  period using [Inspektor Gadget (`ig`)](https://github.com/inspektor-gadget/inspektor-gadget),
  which ships by default on AKS nodes.

## Usage

```bash
# Run a verify check on all nodes
kubectl aks check verify <check-name>

# Run a verify check on a specific node
kubectl aks check verify <check-name> --node <node-name>

# Run a trace check for 30 seconds (default: 10s)
kubectl aks check trace <check-name> --duration 30
```

## Available Checks

### Verify Checks

| Name | Description |
|------|-------------|
| `apiserver-connectivity` | Check connectivity between the node and the Kubernetes API Server |
| `dns-resolution` | Check if the node can resolve required AKS FQDNs (mcr.microsoft.com, login.microsoftonline.com, etc.) |
| `disk-pressure` | Check disk usage and inode exhaustion on the node (>85% threshold) |
| `oom-events` | Check for recent OOM kill events on the node |
| `process-health` | Check that critical node processes (kubelet, containerd) are running |

### Trace Checks

Trace checks use `ig` (Inspektor Gadget) to observe node-level events in real
time. They require the `--duration` flag to specify how long to observe.

| Name         | Description |
|--------------|-------------|
| `dns-failed` | Trace DNS queries reporting failures (NXDOMAIN, SERVFAIL, etc.) |
| `dns-slow`   | Trace DNS queries taking longer than 500ms |
| `tcp-drops`  | Trace TCP packet losses |
| `tcp-retrans` | Trace TCP retransmissions |

## Examples

### Verify DNS resolution on all nodes

```bash
kubectl aks check verify dns-resolution
```

```
=== aks-nodepool1-vmss000000 ===
✔ DNS resolution: all 5 required FQDNs resolved successfully
FQDNs tried: mcr.microsoft.com, eastus.data.mcr.microsoft.com, login.microsoftonline.com, packages.microsoft.com, packages.aks.azure.com

=== aks-nodepool1-vmss000001 ===
✔ DNS resolution: all 5 required FQDNs resolved successfully
FQDNs tried: mcr.microsoft.com, eastus.data.mcr.microsoft.com, login.microsoftonline.com, packages.microsoft.com, packages.aks.azure.com
```

### Trace failing DNS queries for 30 seconds

```bash
kubectl aks check trace failed-dns --duration 30
```

```
✗ DNS trace: 2 DNS failure(s) observed
NAME                RCODE      QTYPE  NAMESERVER       POD                        PROCESS
bad.example.com.    NameError  A      168.63.129.16    default/my-pod             curl(1234)
typo.internal.      ServFail   AAAA   168.63.129.16                               python3(5678)
```

### Check for slow DNS queries

```bash
kubectl aks check trace dns-slow --duration 60
```

```
✗ DNS slow trace: 1 slow DNS query/queries observed (>500ms)
NAME                RCODE    QTYPE  NAMESERVER       LATENCY  POD                    PROCESS
slow.service.com.   Success  A      168.63.129.16    1.2s     kube-system/coredns    curl(900)
```

### Check for TCP packet losses

```bash
kubectl aks check trace tcp-drops --duration 20
```

```
✗ TCP drops trace: 1 packet loss(es) detected
SRC                  DST                POD                        PROCESS
10.244.0.160:59344   20.105.36.95:443   kube-system/konnectivity   proxy-agent(33571)
```

### Check for TCP retransmissions

```bash
kubectl aks check trace tcp-retrans --duration 20
```

```
✗ TCP retrans trace: 1 retransmission(s) detected
SRC                  DST                FLAGS     POD                        PROCESS
10.244.0.160:59344   20.105.36.95:443   PSH|ACK   kube-system/konnectivity   proxy-agent(33571)
```

## Writing a New Check

Adding a check requires implementing a single Go interface and calling
`Register()` in an `init()` function. The framework handles runtime selection,
node targeting, parallel fan-out, and result formatting.

### The Check Interface

```go
package check

type Check interface {
    Name() string                              // Subcommand name (e.g. "dns-resolution")
    Description() string                       // One-line description for --help
    Mode() Mode                                // ModeVerify or ModeTrace
    Command() string                           // Shell command to run on the node
    Parse(res *runtime.RunResult) (*Result, error) // Parse command output into a result
}
```

### Example: A Minimal Verify Check

```go
package check

import pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"

func init() {
    Register(&myCheck{})
}

type myCheck struct{}

func (c *myCheck) Name() string        { return "my-check" }
func (c *myCheck) Description() string { return "Verify something on the node" }
func (c *myCheck) Mode() Mode          { return ModeVerify }

func (c *myCheck) Command() string {
    return `some-command --that-runs-on-the-node`
}

func (c *myCheck) Parse(res *pkgruntime.RunResult) (*Result, error) {
    if res.ExitCode != 0 {
        return &Result{
            Success: false,
            Message: "my-check failed",
            Details: res.Stderr,
        }, nil
    }
    return &Result{Success: true, Message: "my-check passed"}, nil
}
```

### Trace Check Commands

For trace checks that use `ig`, embed the `IGCheck` struct to generate the
command string automatically:

```go
func init() {
    Register(newMyTrace())
}

type myTrace struct {
    IGCheck
}

func newMyTrace() *myTrace {
    return &myTrace{
        IGCheck: IGCheck{
            GadgetImage: "my_gadget",
            Filters:     []string{"some=filter"},
        },
    }
}

func (c *myTrace) Name() string        { return "my-trace" }
func (c *myTrace) Description() string { return "Trace something on the node" }
func (c *myTrace) Mode() Mode          { return ModeTrace }
func (c *myTrace) Command() string     { return c.IGCommand() }
```

`IGCheck` supports the following fields:

| Field | Description | Default |
|-------|-------------|---------|
| `GadgetImage` | The gadget to run (e.g. `trace_dns`) | (required) |
| `OutputMode` | Output format (`--output` flag) | `json` |
| `Filters` | Filter expressions joined by comma for `--filter` | (none) |
| `ExtraArgs` | Additional arguments appended after filters | (none) |

The generated command follows the pattern:
```
ig run <GadgetImage> --host --timeout {{.Duration}} --output <OutputMode> [--filter <Filters>] [ExtraArgs] 2>/dev/null || true
```

### Result Structure

```go
type Result struct {
    Success bool   // true = check passed, false = issue detected
    Message string // One-line summary (always shown)
    Details string // Multi-line detail output (shown below the message)
}
```

### Tips for Check Authors

- **Keep commands self-contained** — The command runs on an AKS node via
  `nsenter` or VMSS RunCommand. It should not depend on files from this repo.
- **Use `ig` for traces** — `ig` is pre-installed on AKS nodes. Use
  `--output json` for structured parsing and `--filter` for server-side
  filtering to reduce data volume.
- **Include pod context** — `ig` output includes `k8s.namespace` and
  `k8s.podName`. Use the shared `K8s` struct and `FormatPod()` helper to
  display a `POD` column as `namespace/pod` in results.
- **Handle empty output** — If the command produces no output (e.g., no
  failures found), return a successful result.
- **VMSS output limit** — The Azure RunCommand API truncates output at ~4KB.
  Filter aggressively on the node side to stay within limits.
