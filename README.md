# Microsoft Azure CLI kubectl plugin

`kubectl-az` is a set of commands used to troubleshoot Kubernetes clusters in
Azure.

Going through the following documentation will help you to understand each
available command and which one is the most suitable for your case:

- [run-command](docs/run-command.md)
- [check-apiserver-connectivity](docs/check-apiserver-connectivity.md)

Consider `kubectl-az` expects the cluster to use virtual machine scale sets.
And, commands that allow using `--node` flag requires the Kubernetes API server
to up and running because it is used to retrieve the VMSS instance information
of nodes.

However, in case of issues with the Kubernetes API server, we can retrieve the
VMSS instance information from the [Azure portal](https://portal.azure.com/) and
pass it to the commands using the `--id` flag or separately with the
`--subscription`, `--node-resource-group`, `--vmss` and `--instance-id` flags.

## Install

```bash
$ git clone https://github.com/Azure/kubectl-az.git
$ cd kubectl-az
# Build and copy the resulting binary in $HOME/.local/bin/
$ make install
```

Notice it requires Go version 1.17.

## Usage

```bash
$ kubectl az --help
Microsoft Azure CLI kubectl plugin

Usage:
  kubectl-az [command]

Available Commands:
  check-apiserver-connectivity  Check connectivity between the nodes and the Kubernetes API Server
  completion                    Generate the autocompletion script for the specified shell
  help                          Help about any command
  run-command                   Run a command in a node
  version                       Show version

Flags:
  -h, --help   help for kubectl-az

Use "kubectl-az [command] --help" for more information about a command.
```

It is necessary to sign in to Azure to run any `kubectl-az` command. To do so,
you can use any authentication method provided by the [Azure
CLI](https://github.com/Azure/azure-cli/) using the `az login` command; see
further details
[here](https://docs.microsoft.com/en-us/cli/azure/authenticate-azure-cli).
However, if you do not have the Azure CLI or have not signed in yet,
`kubectl-az` will open the default browser and load the Azure sign-in page where
you need to authenticate.

## Future

- `kubectl-az` does not store the access token as Azure CLI does. Therefore,
  unless you sign in using Azure CLI, `kubectl-az` will request you to
  authenticate with the default web browser at every execution.

## Contributing

This project welcomes contributions and suggestions.  Most contributions require
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
