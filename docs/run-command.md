# Run Command

We can use `run-command` to execute a command on one of the cluster nodes. Using the proper parameters, this command will work regardless of issues on the Kubernetes API server or, in general, in the Kubernetes control plane.

Take into account that all the
[restrictions](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/run-command#restrictions)
of running scripts in an Azure Linux VM also apply here.

## Regardless of the Kubernetes control plane status

When executing `run-command` without passing through Kubernetes, we need to provide the information of the node (VMSS instance) where we want to run the given command. To retrieve such information, we can use the [`config import`](./config.md#importing-configuration) command. Once we got it, we can select the node we want to use and all the subsequent `run-command` commands will be executed on that node.
To try it start by cleaning the current configuration (if any)

```bash
kubectl aks config unset-all
```

Import the nodes information with the cluster information:

```bash
kubectl aks config import --subscription $mySubID --resource-group $myRG --cluster-name $myCluster
```

In case we want to print the imported information, we can use the `show` command:

```bash
kubectl aks config show
```

If importing the nodes information was successful, we should see something like this:

<!--expected_similarity=0.5-->
```
nodes:
    aks-agentpool-12345678-vmss000000:
        instance-id: "0"
        subscription: mySubID
        node-resource-group: myNRG
        vmss: myVMSS
    aks-agentpool-12345678-vmss000001:
        instance-id: "1"
        [...]
    aks-agentpool-12345678-vmss000002:
        instance-id: "2"
        [...]
```

Start using one of above node e.g `aks-agentpool-12345678-vmss000000` we call it `$myNode` here:

```bash
kubectl aks config use-node $myNode
```

Execute the run-command, and it will be automatically executed in `aks-agentpool-12345678-vmss000000`: 

```bash
kubectl aks run-command "ip route"
```

The output should be similar to this:

<!--expected_similarity=0.8-->
```
default via 10.240.0.1 dev eth0 proto dhcp src 10.240.0.4 metric 100
10.240.0.0/16 dev eth0 proto kernel scope link src 10.240.0.4
10.244.2.2 dev calic38a36632c7 scope link
10.244.2.6 dev cali0b155bb80e7 scope link
10.244.2.7 dev cali997a02e57a6 scope link
10.244.2.8 dev calia2f1486fcb5 scope link
10.244.2.9 dev cali221544885dd scope link
10.244.2.10 dev cali8913de1b395 scope link
10.244.2.14 dev cali8eecb1f59c6 scope link
168.63.129.16 via 10.240.0.1 dev eth0 proto dhcp src 10.240.0.4 metric 100
169.254.169.254 via 10.240.0.1 dev eth0 proto dhcp src 10.240.0.4 metric 100
```

If we run another command, it will again be executed in `aks-agentpool-12345678-vmss000000`:

```bash
kubectl aks run-command "hostname"
```

The output should be similar to this:

<!--expected_similarity=0.8-->
```
aks-agentpool-12345678-vmss000000
```

Unset the current node to avoid conflict with flags or environment variables:

```bash 
kubectl aks config unset-current-node
```

On the other side, if we already have the node (VMSS instance) information and we don't want/need to save it locally, we could pass it directly as following:

<!-- TODO: Test following when we have a simple way to get instance information -->
```
kubectl aks run-command "ip route" --id "/subscriptions/$mySubID/resourceGroups/$myNRG/providers/Microsoft.Compute/virtualMachineScaleSets/$myVMSS/virtualmachines/$myInsId"
```

Or using the flags:

```
kubectl aks run-command "ip route" --subscription $mySubID --node-resource-group $myNRG --vmss $myVMSS --instance-id $myInsId
```

## Passing through Kubernetes

If we are debugging a node while the Kubernetes control plane is up and running, we can simply pass the node name to the `run-command` and it will internally retrieve all the data it needs from the API server to execute the command in that node:

```bash
kubectl aks run-command "ip route" --node $myNode
```

In addition, if we need to run multiple commands on a node, we can still use the [`config import`](./config.md#importing-configuration) command to import the information of all the nodes of our cluster, and this time we don't need to pass the cluster information as `run-command` will retrieve it from the API server:

```bash
kubectl aks config import
```

Start using one of the nodes e.g `aks-agentpool-12345678-vmss000000` we call it `$myNode` here:

```bash
kubectl aks config use-node $myNode
```

Execute the run-command, and it will be automatically executed in `aks-agentpool-12345678-vmss000000`:

```bash
kubectl aks run-command "ip route"
```

If we run another command, it will again be executed in `aks-agentpool-12345678-vmss000000`:

```bash
kubectl aks run-command "hostname"
```
