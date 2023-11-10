# Azure Kubernetes Service (AKS) kubectl plugin

`kubectl-aks` is a `kubectl` plugin that provides a set of commands that enable
users to interact with an AKS cluster even when the control plane is not
functioning as expected. For example, users can still use the plugin to debug
their cluster if the API server is not working correctly. This plugin allows
users to perform various tasks, retrieve information, and execute commands
against the cluster nodes, regardless of the control plane's state.

It's important to note that this plugin does not replace the Azure CLI,
[az](https://learn.microsoft.com/en-us/cli/azure/?view=azure-cli-latest).
Instead, it complements it by offering additional commands and providing users
with a kubectl-like experience. In practice, users will use az to create and
delete their AKS cluster, and then use kubectl and kubectl-aks to interact with
and debug it.

Going through the following documentation will help you to understand each
available command and which one is the most suitable for your case:

- [run-command](docs/run-command.md)
- [check-apiserver-connectivity](docs/check-apiserver-connectivity.md)
- [config](docs/config.md)

Consider `kubectl-aks` expects the cluster to use virtual machine scale sets,
which is the case of an AKS cluster. And, the use of the `--node` flag requires
the Kubernetes control plane to up and running, because the VMSS instance
information of the node will be retrieved from the Kubernetes API server.

However, in case of issues with the Kubernetes control plane, you can reuse the
already stored VMSS instance information, see [config](docs/config.md) command.
Or, if it is a cluster you have never used before on that host, you can retrieve
such information from the [Azure portal](https://portal.azure.com/) and pass it
to the commands using the `--id` flag or separately with the `--subscription`,
`--node-resource-group`, `--vmss` and `--instance-id` flags.

## Install

There is multiple ways to install the `kubectl-aks`.

### Using krew

[krew](https://sigs.k8s.io/krew) is the recommended way to install `kubectl-aks`.
You can follow the [krew's
quickstart](https://krew.sigs.k8s.io/docs/user-guide/quickstart/) to install it
and then install `kubectl-aks` by executing the following command:

```bash
kubectl krew install aks
kubectl aks version
```

It can be uninstalled using the following command:

```bash
kubectl krew uninstall aks
```

### Install a specific release

It is possible to download the asset for a given release and platform from the
[releases page](https://github.com/azure/kubectl-aks/releases/), uncompress and
move the `kubectl-aks` executable to any folder in your `$PATH`.

```bash
VERSION=$(curl -s https://api.github.com/repos/azure/kubectl-aks/releases/latest | jq -r .tag_name)
curl -sL https://github.com/azure/kubectl-aks/releases/latest/download/kubectl-aks-linux-amd64-${VERSION}.tar.gz | sudo tar -C ${HOME}/.local/bin -xzf - kubectl-aks
kubectl aks version
```

It can be uninstalled by using the following command:

```bash
rm ${HOME}/.local/bin/kubectl-aks
```

### Compile from source

To build `kubectl-aks` from source, you'll need to have a Golang version 1.17
or higher installed:

```bash
git clone https://github.com/Azure/kubectl-aks.git
cd kubectl-aks
# Build and copy the resulting binary in $HOME/.local/bin/
make install
kubectl aks version
```

It can be uninstalled by using the following command:

```bash
make uninstall
```

## Usage

```
$ kubectl aks --help
Azure Kubernetes Service (AKS) kubectl plugin

Usage:
  kubectl-aks [command]

Available Commands:
  check-apiserver-connectivity Check connectivity between the nodes and the Kubernetes API Server
  completion                   Generate the autocompletion script for the specified shell
  config                       Manage configuration
  help                         Help about any command
  run-command                  Run a command in a node
  version                      Show version

Flags:
  -h, --help   help for kubectl-aks

Use "kubectl-aks [command] --help" for more information about a command.
```

It is necessary to sign in to Azure to run any `kubectl-aks` command. To do so,
you can use any authentication method provided by the [Azure
CLI](https://github.com/Azure/azure-cli/) using the `az login` command; see
further details
[here](https://docs.microsoft.com/en-us/cli/azure/authenticate-azure-cli).
However, if you do not have the Azure CLI or have not signed in yet,
`kubectl-aks` will open the default browser and load the Azure sign-in page where
you need to authenticate.

### Permissions

In order to run `kubectl-aks` commands, the user/service principal must have the permissions to perform the
following [operations](https://learn.microsoft.com/en-us/azure/role-based-access-control/resource-provider-operations):

- Run command on the instances: `Microsoft.Compute/virtualMachineScaleSets/virtualmachines/runCommand/action`
- List Virtual Machine Scale Sets (VMSS): `Microsoft.Compute/virtualMachineScaleSets/virtualMachines/read`
- List Virtual Machine Scale Set Instances (VMSS Instances): `Microsoft.Compute/virtualMachineScaleSets/read`

Normally if you are using [built-in](https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles)
roles e.g Contributor, you should have the above permissions. However, if you are
using [custom roles](https://learn.microsoft.com/en-us/azure/role-based-access-control/custom-roles-portal) for a
service principal, you need to make sure that the permissions are granted.

## Contributing

This project welcomes contributions and suggestions. Most contributions require
you to agree to a Contributor License Agreement (CLA) declaring that you have
the right to, and actually do, grant us the rights to use your contribution. For
details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether
you need to provide a CLA and decorate the PR appropriately (e.g., status check,
comment). Simply follow the instructions provided by the bot. You will only need
to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of
Conduct](https://opensource.microsoft.com/codeofconduct/). For more information
see the [Code of Conduct
FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact
[opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional
questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or
services. Authorized use of Microsoft trademarks or logos is subject to and must
follow [Microsoft's Trademark & Brand
Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must
not cause confusion or imply Microsoft sponsorship. Any use of third-party
trademarks or logos are subject to those third-party's policies.
