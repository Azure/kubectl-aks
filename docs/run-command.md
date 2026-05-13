# Run Command

We can use `run-command` to execute a command on one or more cluster nodes. Using the proper parameters, this command will work regardless of issues on the Kubernetes API server or, in general, in the Kubernetes control plane.

Take into account that all the
[restrictions](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/run-command#restrictions)
of running scripts in an Azure Linux VM also apply here.

## Running on a single node

### Regardless of the Kubernetes control plane status

When executing `run-command` without passing through Kubernetes, we need to provide the information of the node (VMSS instance) where we want to run the given command. To retrieve such information, we can use the [`config import`](./config.md#importing-configuration) command. Once we got it, we can select the node we want to use and all the subsequent `run-command` commands will be executed on that node:

```bash
# Import the nodes information with the cluster information
$ kubectl aks config import --subscription mySubID --resource-group myRG --cluster-name myCluster
$ kubectl aks config use-node aks-agentpool-12345678-vmss000000

# Execute the run-command on the active node
$ kubectl aks run-command "ip route"
default via 10.240.0.1 dev eth0 proto dhcp src 10.240.0.4 metric 100
10.240.0.0/16 dev eth0 proto kernel scope link src 10.240.0.4
...
```

On the other side, if we already have the node (VMSS instance) information and we don't want/need to save it locally, we could pass it directly as following:

```bash
kubectl aks run-command "ip route" --id "/subscriptions/$SUBSCRIPTION/resourceGroups/$NODERESOURCEGROUP/providers/Microsoft.Compute/virtualMachineScaleSets/$VMSS/virtualmachines/$INSTANCEID"
```

```bash
kubectl aks run-command "ip route" --subscription $SUBSCRIPTION --node-resource-group $NODERESOURCEGROUP --vmss $VMSS --instance-id $INSTANCEID
```

### Passing through Kubernetes

If we are debugging a node while the Kubernetes control plane is up and running, we can simply pass the node name to the `run-command` and it will internally retrieve all the data it needs from the API server to execute the command in that node:

```bash
kubectl aks run-command "ip route" --node aks-agentpool-12345678-vmss000000
```

### Using kube-api runtime

If you prefer to run commands via a privileged debug pod (instead of the Azure VMSS RunCommand API), you can use the `--runtime kube-api` flag. This creates an ephemeral pod on the target node with `nsenter` for full host-level access:

```bash
kubectl aks run-command "ip route" --node aks-agentpool-12345678-vmss000000 --runtime kube-api
```

You can also specify a custom container image for the debug pod:

```bash
kubectl aks run-command "hostname" --node my-node --runtime kube-api --debug-image alpine:latest
```

**Note:** The `kube-api` runtime requires a functioning Kubernetes API server and only needs the `--node` flag (not VMSS instance details).

## Running across an entire cluster (fan-out)

When you have a cluster configured, you can run a command across **all nodes in
the cluster** in parallel. This is useful for collecting diagnostics, checking
uptime, or running any command cluster-wide.

### Interactive selection

If no `--node` or VMSS instance flags are provided and a cluster is set,
`kubectl-aks` will present an interactive prompt:

```bash
$ kubectl aks run-command "uptime"
? Select node in myCluster:
  ▸ (all nodes in cluster)
    aks-nodepool1-12345678-vmss000000
    aks-nodepool1-12345678-vmss000001
    aks-nodepool1-12345678-vmss000002
```

Select **"(all nodes in cluster)"** to execute on every node in parallel, or
pick a specific node.

### Using the --cluster-name flag

You can also use the `--cluster-name` flag to run across all nodes in a cluster
without interactive selection:

```bash
$ kubectl aks run-command "uptime" --cluster-name myCluster
=== aks-nodepool1-12345678-vmss000000 ===
 12:30:00 up 5 days, ...

=== aks-nodepool1-12345678-vmss000001 ===
 12:30:00 up 5 days, ...

=== aks-nodepool1-12345678-vmss000002 ===
 12:30:00 up 5 days, ...
```

Fan-out execution respects the `--runtime` flag, so you can combine it with
`--runtime kube-api` to use the Kubernetes debug pod approach across all nodes:

```bash
kubectl aks run-command "hostname" --cluster-name myCluster --runtime kube-api
```

## Saving configuration for repeated use

If we need to run multiple commands on a node, we can still use the [`config import`](./config.md#importing-configuration) command to import the information of all the nodes of our cluster:

```bash
# Import the nodes information via kubeconfig
kubectl aks config import --runtime kube-api

# Start using one of the nodes
kubectl aks config use-node aks-agentpool-12345678-vmss000000

# Execute commands on the active node
kubectl aks run-command "ip route"
kubectl aks run-command "hostname"
```
