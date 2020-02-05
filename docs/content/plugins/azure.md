---
title: azure plugin
description: Integrate Porter with Azure Cloud
---

<img src="/images/mixins/azure.png" class="mixin-logo" style="width: 300px"/>

Integrate Porter with Azure Cloud.

Source: https://github.com/deislabs/porter-azure-plugins

## Install or Upgrade

```
porter plugin install azure
```

## Plugin Configuration

### Storage

Storage plugins allow Porter to store data, such as claims and credentials, in the Azure's cloud.

#### Blob

The `azure.blob` plugin stores data in Azure Blob Storage. 

1. Open, or create, `~/.porter/config.toml`.
1. Add the following line to activate the Azure blob storage plugin:

    ```toml
    default-storage-plugin = "azure.blob"
    ```

1. [Create a storage account][account]
1. [Create a container][container] named `porter`.
1. [Copy the connection string][connstring] for the storage account. Then set it as an environment variable named 
    `AZURE_STORAGE_CONNECTION_STRING`.

### Secrets

Secrets plugins allow Porter to inject secrets into credential sets.

For example, if your team has a shared key vault with a database password, you
can use the keyvault plugin to inject it as a credential when you install a bundle.

#### Key Vault

The `azure.keyvault` plugin resolves credentials against secrets in Azure Key Vault.

1. Open, or create, `~/.porter/config.toml`
1. Add the following lines to activate the Azure keyvault secrets plugin:

    ```toml
    default-secrets = "mysecrets"
    
    [[secrets]]
    name = "mysecrets"
    plugin = "azure.keyvault"
    
    [secrets.config]
    vault = "myvault"
    ```
1. [Create a key vault][keyvault] and set the vault name in the config with name of the vault.
1. [Create a service principal][sp] and create an Access Policy on the vault giving Get and List secret permissions.
1. Using credentials for the service principal set the environment variables: `AZURE_TENANT_ID`,`AZURE_CLIENT_ID`,  and `AZURE_CLIENT_SECRET`.

[account]: https://docs.microsoft.com/en-us/azure/storage/common/storage-quickstart-create-account?tabs=azure-portal
[container]: https://docs.microsoft.com/en-us/azure/storage/blobs/storage-quickstart-blobs-portal#create-a-container
[connstring]: https://docs.microsoft.com/en-us/azure/storage/common/storage-configure-connection-string?toc=%2fazure%2fstorage%2fblobs%2ftoc.json#view-and-copy-a-connection-string
[keyvault]: https://docs.microsoft.com/en-us/azure/key-vault/quick-create-portal#create-a-vault
[sp]: https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal