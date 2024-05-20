# Contributing to kubectl-aks

Welcome to the kubectl-aks repository! This document describes different ways to facilitate your contribution to the project.

If you would like to become a contributor to this project (or any other open source Microsoft project), see how to [Get Involved](https://opensource.microsoft.com/collaborate/).

## Testing

### integration tests

The integration test needs an AKS cluster to run against. After you have [created one](https://learn.microsoft.com/en-us/azure/aks/learn/quick-kubernetes-deploy-portal?tabs=azure-cli) and set up the access, use following command to run the integration test:

```
make integration-test
```

You will need to set `AZURE_RESOURCE_GROUP` and `AZURE_CLUSTER_NAME` environment variables to specify the AKS cluster to run against.

### documentation tests

The documentation tests are used to validate the documentation examples. You can run the documentation tests with the following command:

```
make documentation-test-readme
make documentation-test-commands
```

or you can select a specific markdown file of a command using:

```
DOCUMENTATION_TEST_FILES=./docs/run-command.md make documentation-test-commands
```

An INI file with all the required variables needs to be created to run the documentation tests. A sample INI file for [docs/run-command.md](docs/run-command.md)
is available [here](docs/run-command.ini.sample).
