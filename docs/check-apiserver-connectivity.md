# Check API Server Connectivity

We can use `check-apiserver-connectivity` to verify the connectivity between the
nodes and the Kubernetes API server by executing `kubectl version` from the node
itself. This command will try to contact the API server to get the Kubernetes
version it is running, which is enough to verify the connectivity. We have to
consider that `kubectl` uses the URL of the API server available in the
`kubeconfig` file and not directly the IP address. It means that this
connectivity check requires the DNS to be working correctly to succeed.

```bash
$ kubectl get nodes
NAME                                STATUS   ROLES   AGE   VERSION
aks-agentpool-27170680-vmss000000   Ready    agent   11d   v1.22.4
aks-agentpool-27170680-vmss000001   Ready    agent   11d   v1.22.4
aks-agentpool-27170680-vmss000002   Ready    agent   11d   v1.22.4

$ kubectl aks check-apiserver-connectivity --node aks-agentpool-27170680-vmss000000
Connectivity check: succeeded
```

Notice that when we use the `--node` flags, the command
`check-apiserver-connectivity` will need to resolve such node name to the VMSS
instance information using the API server. So, if we suspect there might be an
issue on the API server itself, we can
[import](../docs/config.md#importing-configuration) such information with the
`config` command, as it can the Azure API to do it:

```bash
# Providing the cluster information so that the node information is retrieved using the Azure API
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
$ kubectl aks config use-node aks-agentpool-12345678-vmss000000

# Execute the check-apiserver-connectivity, and it will be automatically executed in aks-agentpool-12345678-vmss000000
$ kubectl aks check-apiserver-connectivity
```

Or, if we already have the VMSS instance information, we can pass it directly:

```bash
kubectl aks check-apiserver-connectivity --id "/subscriptions/$SUBSCRIPTION/resourceGroups/$NODERESOURCEGROUP/providers/Microsoft.Compute/virtualMachineScaleSets/$VMSS/virtualmachines/$INSTANCEID"
```

```bash
kubectl aks check-apiserver-connectivity --subscription $SUBSCRIPTION --node-resource-group $NODERESOURCEGROUP --vmss $VMSS --instance-id $INSTANCEID
```

For debugging purposes, we can use the flag `-v`/`--verbose` to have further
details about the command that is being executed in the nodes to check
connectivity:

```bash
$ kubectl aks check-apiserver-connectivity --node aks-agentpool-27170680-vmss000001 -v
Command: kubectl --kubeconfig /var/lib/kubelet/kubeconfig version > /dev/null; echo $?
Virtual Machine Scale Set VM:
{
  "SubscriptionID": "MySub",
  "NodeResourceGroup": "MyNodeRG",
  "VMScaleSet": "MyVMSS",
  "InstanceID": "X"
}

|

Response:
{
  "value": [
    {
      "code": "ProvisioningState/succeeded",
      "displayStatus": "Provisioning succeeded",
      "level": "Info",
      "message": "Enable succeeded: \n[stdout]\n0\n\n[stderr]\n"
    }
  ]
}

Connectivity check: succeeded
```

Given that the `check-apiserver-connectivity` command checks the connectivity by
running a command on the nodes, all the
[restrictions](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/run-command#restrictions)
of running scripts in an Azure Linux VM also apply here.
