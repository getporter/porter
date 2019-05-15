# Porter Examples

This directory contains several example Porter bundles that demonstrate various Porter capabilities.

## Azure Ark

This example creates an Azure storage account using the `azure` and then uses the `helm` mixin to install [Valero](https://github.com/heptio/velero) into an existing Kubernetes cluster. In order to use install this example, you will need an Azure account and a [Service Principal](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest). You'll also need a Kubernetes cluster in Azure, we recommend using AKS.

## Azure MySQL WordPress

This example creates an Azure MySQL database using the `azure` mixin and then uses the `helm` mixin to install WordPress into an existing Kubernetes cluster. In order to use install this example, you will need an Azure account and a [Service Principal](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest). You'll also need a Kubernetes cluster in Azure, we recommend using AKS.

## AKS Spring Music

This example represents an advanced use case for Porter. First, the example showcases the use of a base Dockerfile. This allows you to customize the resulting invocation image by installing software not installed by mixins or by adding any other resources needed for your bundle. Next, it uses the `azure` mixin to create an AKS cluster. Next, it uses the `exec` mixin to obtain the kubeconfig for the AKS cluster and the `kubernetes` mixin to apply RBAC policies in support of Tiller. Then, it uses the `exec` mixin to initialize Helm's Tiller component. Once that has been completed, it uses the `azure` mixin to create a CosmosDB instance with MongoDB API compatability. Finally, it uses the `helm` mixin to install the Spring Music showcase application. This bundle requires an Azure account a [Service Principal](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest).

