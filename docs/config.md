# Config

We can use the `config` command to configure node information for `kubectl-az` commands.
This allows ease of switching between different nodes and persisting the configuration. Nodes can be configured
either specifying the `--node` or `--resource-id` or VMSS instance information(`--subscription`, `--node-resource-group`, `--vmss`, `--instance-id`).
All three options are mutually exclusive:

```bash
$ kubectl az config --help
Manage configuration

Usage:
  kubectl-az config [command]

Available Commands:
  set-node           Set a given node in the configuration
  show               Show the configuration
  unset-all          Unset all nodes in the configuration
  unset-current-node Unset the current node in the configuration
  unset-node         Unset a given node in the configuration
  use-node           Set the current node in the configuration

Flags:
  -h, --help      help for config
  -v, --verbose   Verbose output.

Use "kubectl-az config [command] --help" for more information about a command.
```

As an example, we set a couple of nodes in the configuration (using VMSS instance information) and then switch between them:

```bash
$ kubectl az config set-node node1 --subscription mySubID --node-resource-group myRG --vmss myVMSS --instance-id myInstanceID1
$ kubectl az config set-node node2 --subscription mySubID --node-resource-group myRG --vmss myVMSS --instance-id myInstanceID2
$ kubectl az show
nodes:
    node1:
        instance-id: myInstanceID1
        node-resource-group: myRG
        subscription: mySubID
        vmss: myVMSS
    node2:
        instance-id: myInstanceID2
        node-resource-group: myRG
        subscription: mySubID
        vmss: myVMSS

$ kubectl az config use-node node1
$ kubectl az check-apiserver-connectivity

$ kubectl az config use-node node2
$ kubectl az check-apiserver-connectivity
```

There is also an option to unset node information from the configuration using
the `unset-node`/`unset-all`/`unset-current-node` commands.

## Precedence of configuration

Apart from the configuration file, we can also use the flags and environment variables to
pass the node information to the commands. The precedence of the configuration is the following:

1. Flags
2. Environment variables
3. Configuration file

Using the flags:

```bash
$ kubectl az check-apiserver-connectivity --node aks-agentpool-77471288-vmss000013
```

or using the environment variables:

- `KUBECTL_AZ_NODE`
- `KUBECTL_AZ_RESOURCE_ID`
- `KUBECTL_AZ_SUBSCRIPTION`, `KUBECTL_AZ_NODE_RESOURCE_GROUP`, `KUBECTL_AZ_VMSS` and `KUBECTL_AZ_INSTANCE_ID`

```bash
$ KUBECTL_AZ_NODE=aks-agentpool-77471288-vmss000013 kubectl az check-apiserver-connectivity
```