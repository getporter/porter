# Using Porter with Azure and Terraform

This bundle provides an example of how you can use Porter to build Terraform-based bundles. The example provided here will create Azure CosmosDB and Azure EventHubs objects using Terraform configurations and the [porter-terraform](https://github.com/deislabs/porter-terraform/) mixin. This sample also shows how the Terraform mixin can be used with other mixins, in this case the ARM mixin. The ARM mixin is first used to create an Azure storage account that will be used to configure the Terraform `azurerm` backend. It is possible to build bundles using just the [porter-terraform](https://github.com/deislabs/porter-terraform) mixin, but this example shows you how to use outputs between steps as well.

## Setup

This bundle will create resources in Azure. In order to do this, you'll first need to [create an Azure account](https://azure.microsoft.com/en-us/free/) if you don't already have one.

The bundle will use an Azure [Service Principal](https://docs.microsoft.com/en-us/azure/active-directory/develop/app-objects-and-service-principals) in order to authenticate with Azure. Once you have an account, create a Service Principal for use with the bundle. You can do this via the Azure portal, or via the Azure CLI:

1. Create a service principal with the Azure CLI:
    ```console
    az ad sp create-for-rbac --name porterform -o table
    ```
1. Save the values from the command output in environment variables:

    **Bash**
    ```console
    export AZURE_TENANT_ID=<Tenant>
    export AZURE_CLIENT_ID=<AppId>
    export AZURE_CLIENT_SECRET=<Password>
    ```

    **PowerShell**
    ```console
    $env:AZURE_TENANT_ID = "<Tenant>"
    $env:AZURE_CLIENT_ID = "<AppId>"
    $env:AZURE_CLIENT_SECRET = "<Password>"
    ```

You will also need to have [Porter](https://porter.sh/install/) and [Docker](https://docs.docker.com/v17.12/install/) installed before you proceed.

## Overview

If you examine the contents of this example, you will see a `porter.yaml` file and a `terraform` directory. The `porter.yaml` is the bundle definition that will be used to build and run the bundle. Please see the [porter.yaml](porter.yaml) for the contents of the file. In this file, you will find a few credential definitions:

```yaml
credentials:
  - name: subscription_id
    env: AZURE_SUBSCRIPTION_ID

  - name: tenant_id
    env: AZURE_TENANT_ID

  - name: client_id
    env: AZURE_CLIENT_ID

  - name: client_secret
    env: AZURE_CLIENT_SECRET
```

These credentials will correspond to the Azure Service Principal you created above. You'll also notice some parameter definitions:

```yaml
parameters:
  - name: location
    type: string
    default: "EastUS"

  - name: resource_group_name
    type: string
    default: "porter-terraform"

  - name: storage_account_name
    type: string
    default: "porterstorage"

  - name: storage_container_name
    type: string
    default: "portertf"

  - name: storage_rg
    type: string
    default: "porter-storage"

  - name: database-name
    type: string
    default: "porter-terraform"
```

These represent the values that can be provided at runtime when installing the bundle.

Finally, please note how the `porter-terraform` mixin is used:

```yaml
- terraform:
      description: "Create Azure CosmosDB and Event Hubs"
      input: false
      backendConfig:
        key: "{{ bundle.name }}.tfstate"
        storage_account_name: "{{ bundle.parameters.storage_account_name }}"
        container_name: "{{ bundle.parameters.storage_container_name }}"
        access_key: "{{ bundle.outputs.STORAGE_ACCOUNT_KEY }}"
      vars:
        subscription_id: "{{bundle.credentials.subscription_id}}"
        tenant_id: "{{bundle.credentials.tenant_id}}"
        client_id: "{{bundle.credentials.client_id}}"
        client_secret: "{{bundle.credentials.client_secret}}"
        database_name: "{{bundle.parameters.database-name}}"
        resource_group_name: "{{bundle.parameters.resource_group_name}}"
        resource_group_location: "{{bundle.parameters.location}}"
      outputs:
      - name: cosmos-db-uri
      - name: eventhubs_connection_string
```

This mixin step uses both parameters and credentials, defined above, and declares two output values. The source of these output files is defined in the Terraform configuration.

The `terraform` directory contains the Terraform configuration files that will be used by the `porter-terraform` mixin:

```bash
ls -l terraform/
total 40
-rw-r--r--  1 jeremyrickard  staff  1023 Jul  2 08:32 cosmos-db.tf
-rw-r--r--  1 jeremyrickard  staff   622 Jul  2 08:30 eventhubs.tf
-rw-r--r--  1 jeremyrickard  staff   290 Jul  3 10:43 main.tf
-rw-r--r--  1 jeremyrickard  staff   286 Jul  3 17:09 outputs.tf
-rw-r--r--  1 jeremyrickard  staff   325 Jul  2 08:11 variables.tf
```

The `main.tf` file configures the `azurerm` provider, while the `outputs.tf` defines the outputs we will capture from the Terraform step. `cosmos-db.tf` and `eventhubs.tf` contain the declarations for the infrastructure we will create. Finally, `variables.tf` defines a set of variables used throughout the files. These correspond to the parameters in our `porter.yaml` above.

## Building The Bundle

In order to use this bundle, you'll first need to build it. This is done with the `porter` command line tool. Ensure that your working directory is set to this example directory before proceeding, then run the following command:

```
porter build
```

Once this command has finished, you will see some additional resources in your working directory: a `Dockerfile` and a `.cnab` directory. The `Dockerfile` was generated by Porter and the `porter-terraform` mixin:

```bash
$ more Dockerfile
FROM debian:stretch

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates

# exec mixin has no buildtime dependencies

ENV TERRAFORM_VERSION=0.12.17
RUN apt-get update && apt-get install -y wget unzip && \
 wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
 unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin
COPY . $BUNDLE_DIR
RUN cd /cnab/app/terraform && terraform init -backend=false


COPY . $BUNDLE_DIR
RUN rm -fr $BUNDLE_DIR/.cnab
COPY .cnab /cnab
COPY porter.yaml $BUNDLE_DIR/porter.yaml
WORKDIR $BUNDLE_DIR
CMD ["/cnab/app/run"]
```

The Dockerfile contains the necessary instructions to install Terraform within our bundle's invocation image.

## Generate a Credential Set

Before you can install the bundle, you'll need to generate a credential set. This credential set will map your local service principal environment variables into the destinations defined in the `porter.yaml`. You can generate the credential set by running `porter credentials generate`. This will prompt you for the source for each credential defined in the bundle:

```bash
$ porter credentials generate
Generating new credential azure-terraform from bundle azure-terraform
==> 4 credentials required for bundle azure-terraform
? How would you like to set credential "client_id"  [Use arrows to move, space to select, type to filter]
  specific value
> environment variable
  file path
  shell command
```

For each of the credentials, provide the corresponding environment variable that you set above. For example, for `client_id`, select `environment variable` and provide the value `AZURE_CLIENT_ID`.

```bash
$ porter credentials generate
Generating new credential azure-terraform from bundle azure-terraform
==> 4 credentials required for bundle azure-terraform
? How would you like to set credential "client_id" environment variable
? Enter the environment variable that will be used to set credential "client_id" AZURE_CLIENT_ID
? How would you like to set credential "client_secret" environment variable
? Enter the environment variable that will be used to set credential "client_secret" AZURE_CLIENT_SECRET
? How would you like to set credential "subscription_id" environment variable
? Enter the environment variable that will be used to set credential "subscription_id" AZURE_SUBSCRIPTION_ID
? How would you like to set credential "tenant_id" environment variable
? Enter the environment variable that will be used to set credential "tenant_id" AZURE_TENANT_ID
Saving credential to /Users/jeremyrickard/.porter/credentials/azure-terraform.yaml
```

## Installing the Bundle

Once you have built the bundle and generated a credential set, you're ready to install the bundle! To do that, you'll use the `porter install` command:

```bash
$ porter install -c azure-terraform
installing azure-terraform...
executing install action from azure-terraform (bundle instance: azure-terraform)
Create an Azure Storage Account
Starting deployment operations...
Finished deployment operations...
Emit the key in base64 encoded form
Here is a the storage account key (base64 encoded) ==> cFNZNExabEg1eGkzSkgrdU5HcFZyek94WmEyYXRRa1Z6WWtFVjZGamg5aU5wcjRVVjROVFBmSXJH
UXNpTVpLQS9FYWVWanF1WkhhMFg5TE9IMERRY2c9PQo=
Create Azure CosmosDB and Event Hubs
Initializing Terraform...
/usr/bin/terraform terraform init -backend=true -backend-config=access_key=******* -backend-config=container_name=portertf -backend-config=key=azure-terraform.tfstate -backend-config=storage_account_name=porterstorage -reconfigure

Initializing the backend...

Successfully configured the backend "azurerm"! Terraform will automatically
use this backend unless the backend configuration changes.

Initializing provider plugins...

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
/usr/bin/terraform terraform apply -auto-approve -input=false -var client_id=******* -var client_secret=******* -var database_name=porter-terraform -var resource_group_location=EastUS -var resource_group_name=porter-terraform -var subscription_id=******* -var tenant_id=*******
Acquiring state lock. This may take a few moments...
azurerm_resource_group.rg: Creating...

< OUTPUT TRUNCATED>
```

Installing the bundle will take some amount of time. As it is running you should see an output line like this:

```
Here is a the storage account key (base64 encoded) ==> <SOME STRING>
```

This is the account key that you'll need to later uninstall the bundle. It will be base64 encoded, so you'll need to decode it before using it to uninstall the bundle:

```bash
$ echo <SOMESTRING> | base64 -D
```

You can also obtain the value from the Azure portal.

## Uninstalling the Bundle

When you're ready to uninstall the bundle, simply run the `porter uninstall` command and provide the storage account key:

```bash
$ porter uninstall -c azure-terraform --param tf_storage_account_key=%%YOUR KEY VALUE%%
uninstalling azure-terraform...
executing uninstall action from azure-terraform (bundle instance: azure-terraform)
Remove Azure CosmosDB and Event Hubs
Initializing Terraform...
/usr/bin/terraform terraform init -backend=true -backend-config=access_key=******* -backend-config=container_name=portertf -backend-config=key=azure-terraform.tfstate -backend-config=storage_account_name=porterstorage -reconfigure

Initializing the backend...

Successfully configured the backend "azurerm"! Terraform will automatically
use this backend unless the backend configuration changes.

Initializing provider plugins...

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
/usr/bin/terraform terraform destroy -auto-approve -var client_id=******* -var client_secret=******* -var database_name=porter-terraform -var resource_group_location=EastUS -var resource_group_name=porter-terraform -var subscription_id=******* -var tenant_id=*******
Acquiring state lock. This may take a few moments...

< OUTPUT TRUNCATED>
```

This will take a number of minutes to finish, but when complete the resources will be removed from your account.
