# Porter Examples

This directory contains several example Porter bundles that demonstrate various Porter capabilities.

## AKS Spring Music

This example represents an advanced use case for Porter. First, the example showcases the use of a base Dockerfile. This allows you to customize the resulting invocation image by installing software not installed by mixins or by adding any other resources needed for your bundle. Next, it uses the `azure` mixin to create an AKS cluster. Next, it uses the `exec` mixin to obtain the kubeconfig for the AKS cluster and the `kubernetes` mixin to apply RBAC policies in support of Tiller. Then, it uses the `exec` mixin to initialize Helm's Tiller component. Once that has been completed, it uses the `azure` mixin to create a CosmosDB instance with MongoDB API compatability. Finally, it uses the `helm` mixin to install the Spring Music showcase application. This bundle requires an Azure account and a [Service Principal](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest).

## Azure Terraform

This bundle provides an example of how you can use Porter to build Terraform-based bundles. The example provided here will create Azure CosmosDB and Azure EventHubs objects using Terraform configurations and the [porter-terraform](https://github.com/deislabs/porter-terraform/) mixin. This sample also shows how the Terraform mixin can be used with other mixins, in this case the Azure mixin. The Azure mixin is first used to create an Azure storage account that will be used to configure the Terraform `azurerm` backend. It is possible to build bundles using just the [porter-terraform](https://github.com/deislabs/porter-terraform) mixin, but this example shows you how to use outputs between steps as well.

## Azure MySQL WordPress

This example creates an Azure MySQL database using the `azure` mixin and then uses the `helm` mixin to install WordPress into an existing Kubernetes cluster. In order to use install this example, you will need an Azure account and a [Service Principal](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest). You'll also need a Kubernetes cluster in Azure, we recommend using AKS.

## Exec Outputs

This bundle demonstrates how to use outputs with the exec mixin. Most mixins are based on the exec mixin, so you can use what you learn here with them.

## Kubernetes

This is a sample Porter bundle that makes use of both the Kubernetes and Exec mixins. The Kubernetes mixin is used to apply Kubernetes manifests to an existing Kubernetes cluster, creating an NGINX deployment and a service. The Kubernetes mixin is also used to produce an output with the value of the service's ClusterIP.  After the `kubernetes` mixin finishes, the `exec` mixin is used to echo the cluster IP of the service that was created.

## GKE Example

This is a sample Porter bundle that makes use of both the Kubernetes and Exec mixins, albeit tailored for installation on a GKE cluster. The Kubernetes mixin is used to apply Kubernetes manifests to an existing Kubernetes cluster, creating an NGINX deployment and a service. The Kubernetes mixin is also used to produce an output with the value of the service's ClusterIP.  After the `kubernetes` mixin finishes, the `exec` mixin is used to echo the cluster IP of the service that was created.

## Service Fabric CLI

This example shows how to install and use the Service Fabric CLI (sfctl) in a bundle.