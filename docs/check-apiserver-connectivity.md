# Check Connectivity

We can use `check-connectivity` to verify that nodes can communicate with the
Kubernetes API server:

```bash
$ kubectl get nodes
NAME                                STATUS   ROLES   AGE   VERSION
aks-agentpool-27170680-vmss000000   Ready    agent   11d   v1.22.4
aks-agentpool-27170680-vmss000001   Ready    agent   11d   v1.22.4
aks-agentpool-27170680-vmss000002   Ready    agent   11d   v1.22.4

$ kubectl az check-apiserver-connectivity --node aks-agentpool-27170680-vmss000000
Running...

Connectivity check: succeeded
```

Or we could also pass directly the VMSS instance information:

```bash
$ kubectl az check-apiserver-connectivity --id "/subscriptions/$SUBSCRIPTION/resourceGroups/$NODERESOURCEGROUP/providers/Microsoft.Compute/virtualMachineScaleSets/$VMSS/virtualmachines/$INSTANCEID"
```

```bash
$ kubectl az check-apiserver-connectivity --subscription $SUBSCRIPTION --node-resource-group $NODERESOURCEGROUP --vmss $VMSS --instance-id $INSTANCEID
```

The `check-connectivity` command verifies the connectivity between the nodes and
the API server by executing the command `kubectl version` from the node itself.
This command will try to contact the API server to get the Kubernetes version it
is running, which is enough to verify the connectivity. We have to consider that
`kubectl` uses the URL of the API server available in the `kubeconfig` file and
not directly the IP address. It means that this connectivity check requires the
DNS to be working correctly to succeed.

We can use the flag `-v`/`--verbose` to have further details about the command
that is being executed in the nodes to check connectivity:

```bash
$ kubectl az check-apiserver-connectivity --node aks-agentpool-27170680-vmss000001 -v
Command: kubectl --kubeconfig /var/lib/kubelet/kubeconfig version > /dev/null; echo $?
Virtual Machine Scale Set VM:
{
  "SubscriptionID": "MySub",
  "NodeResourceGroup": "MyNodeRG",
  "VMScaleSet": "MyVMSS",
  "InstanceID": "X"
}

Running...

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

Given that the `check-connectivity` command checks the connectivity by running a
command on the nodes, all the
[restrictions](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/run-command#restrictions)
of running scripts in an Azure Linux VM also apply here.
