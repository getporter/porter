# Advanced Azure + Terraform Cloud Native Application Bundle Using Porter

This exercise extends the [porter-tf](https://getporter.org/src/workshop/porter-tf)  example in order provide a more complete example of buiding a CNAB that combines both infrastructure and deployment of an application. As in the `porter-tf` example, we will use the `arm` and `terraform` mixins to provision a MySQL database on Azure. We will then use the `arm` mixin with a custom [ARM](https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-group-authoring-templates) template to deploy a notional web service as an Azure Container Instance. This part of the bundle could easily be replaced with deployment to Kubernetes or any other container runtime system, but this exercise will use Azure.

## Prerequisites

In order to complete this exercise, you will need to have a recent Docker Desktop installed or you'll want to use [Play With Docker](https://labs.play-with-docker.com/) and you'll also need a Docker registry that you can push to. If you don't have one, a free DockerHub account will work. To create one of those, please visit https://hub.docker.com/signup.

You'll also need to make sure Porter is [installed](https://getporter.org/install/).

To install the bundle, you'll also need an Azure credential.

## Review the bundle/terraform directory

The `terraform` directory contains a set of Terraform configurations that will utilize the Azure provider to create an Azure MySQL instance and store the TF state file in an Azure storage account. The files in here aren't special for use with CNAB or Porter. If you've completed the from scratch exercise, these are exactly the same!

## Reviw the bundle/arm directory

## Review the bundle/porter.yaml

Review the `porter.yaml` in the `bundle` diretory see what the bundle will do with the `terraform` and `ARM` manifests. An important aspect to note in this bundle is that each step produces an output. The `azure` mixin is first used to create an Azure storage account, resulting in a storage account key. Next, that storage account key is used as a parameter to the `terraform` mixin, which in turn produces a FQDN for the MySQL instance. This FQDN is used as a parameter for the ARM template deployed with the `azure` mixin. This last step produces an IP_ADDRESS for the container. Both the IP_ADDRESS and the STORAGE_ACCOUNT_KEY are further exposed as bundle outputs. These values can be obtained after installing the bundle.

Finally, the `porter.yaml` also defines an `imageMap`. In this section, you can declare any images that will be used in addition to the invocation image. These are used for digest validation and for creation of thick bundles.

## Update the porter.yaml

Now, update the `porter.yaml` and change the following value:

```
registry: getporter
```

Change the registry to point to your own Docker registry. For example, if my Docker user name is `jeremyrickard`, I'd change that these lines to:

```
registry: jeremyrickard
```

## Build The Bundle!

Now that you have modified the `porter.yaml`, the next step is to build the bundle. To do that, run the following command.

```
porter build
```

That's it! Porter automates all the things that were required in our manual step. Once this is complete, you can review the generated `Dockerfile`, along with the `bundle.json` in the `.cnab` directory. Now, you can install the bundle using Porter.

## Install the Bundle

In order to install this bundle, you will need Azure credentials in the form of a Service Principal. If you already have a service principal, skip the next section.

### Generating a Service Principal

```
az ad sp create-for-rbac --name porter-workshop -o table
```

Once you run this command, it will output a table similar to this:

```
AppId                                 DisplayName         Name                       Password                              Tenant
------------------------------------  ------------------  -------------------------  ------------------------------------  ------------------------------------
<some value>                          porter-workshop     http://porter-workshop      <some value>                            <some value>
```

Copy these values and move on to setting up your environment variables.

### Setting Environment Variables

You'll need the following Service Principal information, along with an Azure Subscription ID:

- Client ID (also called AppId)
- Client Secret (also called Password)
- Tenant Id (also called Tenant)

These will need to be in a set of environment variables for use in generating a CNAB credential set. Set them like this:

```
export AZURE_CLIENT_ID=<CLIENT_ID>
export AZURE_TENANT_ID=<TENANT_ID>
export AZURE_CLIENT_SECRET=<CLIENT_SECRET>
export AZURE_SUBSCRIPTION_ID=<SUBSCRIPTION_ID>
```

### Generate a CNAB CNAB credential set

A CNAB defines one or more `credentials` in the `bundle.json`. In this exercise, we defined these in our `porter.yaml` and the resulting `bundle.json` contains the credential definitions. Before we can install the bundle, we need to create something called a `credential set` that specifies how to map our Service Principal information into these `credentials`. We'll use Porter to do that:

```
porter credentials generate
```

This command will generate a new `credential set` that maps our environment variables to the `credential` in the bundle. Now we're ready to install the bundle.

### Install the Bundle

Now, you're ready to install the bundle. Replace `<your-name>` with a username like `carolynvs`.

```
porter install -c workshop-tf-aci \
    --param server_name=<your-name>sql \
    --param database_name=testworkshop \
    --param backend_storage_account=<your-name>storagetfaci \
    --param backend_storage_container=<your-name>-workshop-tf-aci \
    --param backend_storage_resource_group=<your-name>-workshop-tf-aci
```

### View The Outputs

Once the bundle has been installed, you can use `porter installation show` to see the outputs:

```
$ porter installation show
Name: workshop-tf-aci
Created: 2 minutes ago
Modified: 4 seconds ago
Last Action: install
Last Status: success

Outputs:
-----------------------------------------------
  Name        Type    Value
-----------------------------------------------
  IP_ADDRESS           string  40.88.49.175
  STORAGE_ACCOUNT_KEY  string  <your storage account key>
```

Note that sensitive outputs (`STORAGE_ACCOUNT_KEY` in this example) are replaced by their runtime path
in the default output format (`-o table`), but their values can be seen via `-o json` or `-o yaml`.

This is the IP address of the new ACI container. You can test it out now with `curl`:

```
$ curl http://20.42.26.66:8080
Hello, I'm a webserver that connects to jrrserver981.mysql.database.azure.com
```

You'll want to replace that IP address with what is shown in the output.
