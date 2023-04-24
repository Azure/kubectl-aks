# Run Command

We can use `run-command` to execute a command on one of the cluster nodes:

```bash
$ kubectl get nodes
NAME                                STATUS   ROLES   AGE   VERSION
aks-agentpool-27170680-vmss000000   Ready    agent   11d   v1.22.4
aks-agentpool-27170680-vmss000001   Ready    agent   11d   v1.22.4
aks-agentpool-27170680-vmss000002   Ready    agent   11d   v1.22.4

$ kubectl aks run-command "ip route" --node aks-agentpool-27170680-vmss000000
Running...

[stdout]
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

[stderr]
```

Another example when requested command prints information to the standard error:
```
$ kubectl aks run-command "cat non-existent-file" --node aks-agentpool-27170680-vmss000000
Running...

[stdout]

[stderr]
cat: non-existent-file: No such file or directory
```

Or we could also pass directly the VMSS instance information:

```bash
$ kubectl aks run-command "ip route" --id "/subscriptions/$SUBSCRIPTION/resourceGroups/$NODERESOURCEGROUP/providers/Microsoft.Compute/virtualMachineScaleSets/$VMSS/virtualmachines/$INSTANCEID"
```

```bash
$ kubectl aks run-command "ip route" --subscription $SUBSCRIPTION --node-resource-group $NODERESOURCEGROUP --vmss $VMSS --instance-id $INSTANCEID
```

Take into account that all the
[restrictions](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/run-command#restrictions)
of running scripts in an Azure Linux VM also apply here.