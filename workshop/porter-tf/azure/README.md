# Build an Azure + Terraform Cloud Native Application Bundle Using Porter

This exercise will show you how to build a Terraform based Cloud Native Application Bundle (CNAB) for Azure using Porter. You should complete this after completing the exercise from Scratch in order to better connect the dots between CNAB and Porter!

## Prerequisites

In order to complete this exercise, you will need to have a recent Docker Desktop installed or you'll want to use [Play With Docker](https://labs.play-with-docker.com/) and you'll also need a Docker registry that you can push to. If you don't have one, a free DockerHub account will work. To create one of those, please visit https://hub.docker.com/signup.

You'll also need to make sure Porter is [installed](https://getporter.org/install/).

## Review the terraform/ directory

The `terraform` directory contains a set of Terraform configurations that will utilize the Azure provider to create an Azure MySQL instance and store the TF state file in an Azure storage account. The files in here aren't special for use with CNAB or Porter. If you've completed the from scratch exercise, these are exactly the same!

## Review the porter.yaml

If you completed the from scratch exercise, you've seen generally what is required to build a CNAB. First, you need a run tool that knows how to execute all the required tools in the proper order with parameters and credentials properly wired to each tool. Next, you need to build a Dockerfile that will contain that run tool, along with all of the supporting tools and configuration. Finally, you need to generate a `bundle.json` that contains a reference to the invocation image along with definitions of parameters, credentials and outputs for the bundle.

Porter was developed to make this experience much easier for anyone that wants to build a bundle. Porter is driven from a single declarative file, called the `porter.yaml`. In this file, you define what capabilities you want to leverage (mixins), define your parameters, credentials and outputs, and then declare what should happen for each action (i.e. install, upgrade, uninstall).

Review the `porter.yaml` to see what each of these sections looks like.

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

* Client ID (also called AppId)
* Client Secret (also called Password)
* Tenant Id (also called Tenant)

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
porter install -c workshop-tf \
    --param server_name=<your-name>sql \
    --param database_name=testworkshop \
    --param backend_storage_account=<your-name>storagetf \
    --param backend_storage_container=<your-name>-workshop-tf \
    --param backend_storage_resource_group=<your-name>-workshop-tf
```

### View the Outputs

Now that you've installed the bundle, you can view any outputs that were created with the `porter installation show` command.

```
porter installation show workshop-tf
```
