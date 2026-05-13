# Config

We can use the `config` command to configure node and cluster information for
`kubectl-aks` commands. This allows ease of switching between different clusters
and nodes and persisting the configuration.

Nodes are grouped under **clusters**, so that importing multiple AKS clusters
keeps each cluster's nodes organized separately. When you run a command without
specifying a node, `kubectl-aks` will present an interactive prompt to select a
cluster and node.

```bash
$ kubectl aks config --help
Manage configuration

Usage:
  kubectl-aks config [command]

Available Commands:
  import             Import Kubernetes nodes in the configuration
  list-clusters      List all clusters in the configuration
  set-node           Set a given node in the configuration
  show               Show the configuration
  unset-all          Unset all nodes in the configuration
  unset-cluster      Remove a cluster and all its nodes from the configuration
  unset-current-node Unset the current node in the configuration
  unset-node         Unset a given node in the configuration
  use-cluster        Set the current cluster in the configuration
  use-node           Set the current node in the configuration

Flags:
  -h, --help      help for config
  -v, --verbose   Verbose output.

Use "kubectl-aks config [command] --help" for more information about a command.
```

## Cluster-aware configuration

Nodes are stored under clusters in the configuration file. This makes it easy to
manage multiple AKS clusters at once:

```bash
# Import nodes from two different clusters
$ kubectl aks config import --subscription mySubID --resource-group myRG --cluster-name cluster-prod
$ kubectl aks config import --subscription mySubID --resource-group myRG --cluster-name cluster-dev

# List imported clusters
$ kubectl aks config list-clusters
cluster-prod
cluster-dev

# Set the active cluster
$ kubectl aks config use-cluster cluster-prod

# Show the configuration
$ kubectl aks config show
current-cluster: cluster-prod
clusters:
    cluster-prod:
        nodes:
            aks-nodepool1-12345678-vmss000000:
                instance-id: "0"
                subscription: mySubID
                node-resource-group: myNRG
                vmss: myVMSS
            aks-nodepool1-12345678-vmss000001:
                instance-id: "1"
                [...]
    cluster-dev:
        nodes:
            aks-nodepool1-87654321-vmss000000:
                instance-id: "0"
                [...]
```

You can also manually set nodes within a cluster:

```bash
$ kubectl aks config set-node node1 --subscription mySubID --node-resource-group myRG --vmss myVMSS --instance-id myInstanceID1
$ kubectl aks config set-node node2 --id "/subscriptions/mySubID/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachineScaleSets/myVMSS/virtualmachines/myInstanceID2"
```

### Switching between clusters and nodes

```bash
# Switch the active cluster
$ kubectl aks config use-cluster cluster-prod

# Switch the active node within the current cluster
$ kubectl aks config use-node aks-nodepool1-12345678-vmss000000

# Run a command on the active node
$ kubectl aks run-command "hostname"
```

### Removing clusters and nodes

```bash
# Remove a specific cluster and all its nodes
$ kubectl aks config unset-cluster cluster-dev

# Unset the current node (but keep it in the config)
$ kubectl aks config unset-current-node

# Remove all configuration
$ kubectl aks config unset-all
```

## Interactive selection

When no `--node` or VMSS instance flags are provided and a cluster is configured,
`kubectl-aks` will present an interactive prompt to select a node:

```bash
$ kubectl aks run-command "uptime"
? Select node in cluster-prod:
  ▸ (all nodes in cluster)
    aks-nodepool1-12345678-vmss000000
    aks-nodepool1-12345678-vmss000001
    aks-nodepool1-12345678-vmss000002
```

Selecting **"(all nodes in cluster)"** will execute the command in parallel
across all nodes in the cluster, displaying results grouped by node:

```
=== aks-nodepool1-12345678-vmss000000 ===
 12:30:00 up 5 days, ...

=== aks-nodepool1-12345678-vmss000001 ===
 12:30:00 up 5 days, ...
```

You can also target a specific cluster with the `--cluster-name` flag:

```bash
kubectl aks run-command "uptime" --cluster-name cluster-prod
```

## Importing configuration

`kubectl-aks` can import node information using the Azure API (default) or the
Kubernetes API via the `config import` command. Imported nodes are automatically
grouped under a cluster name.

### From Azure API (default)

By default, `config import` uses the Azure API. You must provide the
`--subscription`, `--resource-group` and `--cluster-name` flags:

```bash
$ kubectl aks config import --subscription mySubID --resource-group myRG --cluster-name myCluster
$ kubectl aks config show
current-cluster: myCluster
clusters:
    myCluster:
        nodes:
            aks-agentpool-12345678-vmss000000:
                instance-id: "0"
                subscription: mySubID
                node-resource-group: myNRG
                vmss: myVMSS
            [...]
```

### From kubeconfig (--runtime kube-api)

If the Kubernetes API server is available, you can import via kubeconfig by
passing `--runtime kube-api`. The cluster name is auto-detected from the current
kubeconfig context:

```bash
# Create a cluster and get credentials
$ az aks create ...
$ az aks get-credentials ...

# Import nodes via Kubernetes API
$ kubectl aks config import --runtime kube-api
$ kubectl aks config show
current-cluster: my-cluster
clusters:
    my-cluster:
        nodes:
            aks-agentpool-12345678-vmss000000:
                instance-id: "0"
                subscription: mySubID
                node-resource-group: myNRG
                vmss: myVMSS
            [...]
```

## Precedence of configuration

Apart from the configuration file, we can also use the flags and environment variables to
pass the node information to the commands. The precedence of the configuration is the following:

1. Flags (`--node`, `--id`, or VMSS instance flags)
2. Environment variables
3. Interactive cluster/node selection (when a cluster is configured)
4. Configuration file (`current-node`)

Using the flags:

```bash
kubectl aks check-apiserver-connectivity --node aks-agentpool-77471288-vmss000013
```

or using the environment variables:

- `KUBECTL_AKS_NODE`
- `KUBECTL_AKS_RESOURCE_ID`
- `KUBECTL_AKS_SUBSCRIPTION`, `KUBECTL_AKS_NODE_RESOURCE_GROUP`, `KUBECTL_AKS_VMSS` and `KUBECTL_AKS_INSTANCE_ID`

```bash
KUBECTL_AKS_NODE=aks-agentpool-77471288-vmss000013 kubectl aks check-apiserver-connectivity
```

## Legacy configuration

If you have a configuration file from an older version of `kubectl-aks` that
uses the flat node format (top-level `nodes:` without clusters), you will see a
warning recommending you delete the old config and re-import:

```
WARNING: legacy config detected (nodes without clusters). Please delete
~/.kubectl-aks/config.yaml and re-import using 'kubectl aks config import'.
```

The legacy format is still supported for backward compatibility, but the new
cluster-aware format is recommended for managing multiple clusters.
