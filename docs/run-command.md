# Run Command

We can use `run-command` to execute a command on one of the cluster nodes. Using the proper parameters, this command will work regardless of issues on the Kubernetes API server or, in general, in the Kubernetes control plane.

Take into account that all the
[restrictions](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/run-command#restrictions)
of running scripts in an Azure Linux VM also apply here.

## Regardless of the Kubernetes control plane status

When executing `run-command` without passing through Kubernetes, we need to provide the information of the node (VMSS instance) where we want to run the given command. To retrieve such information, we can use the [`config import`](./config.md#importing-configuration) command. Once we got it, we can select the node we want to use and all the subsequent `run-command` commands will be executed on that node:

```bash
# Import the nodes information with the cluster information
$ kubectl aks config import --subscription mySubID --resource-group myRG --cluster-name myCluster
$ kubectl aks config show
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

# Start using one of those nodes
$ kubectl aks use-node aks-agentpool-12345678-vmss000000

# Execute the run-command, and it will be automatically executed in aks-agentpool-12345678-vmss000000
$ kubectl aks run-command "ip route"
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

# Another command that will be still executed in aks-agentpool-12345678-vmss000000
$ kubectl aks run-command "hostname"
aks-agentpool-12345678-vmss000000
```

On the other side, if we already have the node (VMSS instance) information and we don't want/need to save it locally, we could pass it directly as following:

```bash
kubectl aks run-command "ip route" --id "/subscriptions/$SUBSCRIPTION/resourceGroups/$NODERESOURCEGROUP/providers/Microsoft.Compute/virtualMachineScaleSets/$VMSS/virtualmachines/$INSTANCEID"
```

```bash
kubectl aks run-command "ip route" --subscription $SUBSCRIPTION --node-resource-group $NODERESOURCEGROUP --vmss $VMSS --instance-id $INSTANCEID
```

## Passing through Kubernetes

If we are debugging a node while the Kubernetes control plane is up and running, we can simply pass the node name to the `run-command` and it will internally retrieve all the data it needs from the API server to execute the command in that node:

```bash
kubectl aks run-command "ip route" --node aks-agentpool-12345678-vmss000000
```

In addition, if we need to run multiple commands on a node, we can still use the [`config import`](./config.md#importing-configuration) command to import the information of all the nodes of our cluster, and this time we don't need to pass the cluster information as `run-command` will retrieve it from the API server:

```bash
# Import the nodes information from the API server
kubectl aks config import

# Start using one of the nodes
kubectl aks use-node aks-agentpool-12345678-vmss000000

# Execute the run-command, and it will be automatically executed in aks-agentpool-12345678-vmss000000
kubectl aks run-command "ip route"

# Another command that will be still executed in aks-agentpool-12345678-vmss000000
kubectl aks run-command "hostname"
```
