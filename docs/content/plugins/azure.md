---
title: Azure Plugin
description: Integrate Porter with Azure Cloud
---

<img src="/images/plugins/azure.png" class="mixin-logo" style="width: 300px"/>

Integrate Porter with Azure Cloud.

Source: https://github.com/getporter/azure-plugins

## Install or Upgrade

```
porter plugin install azure --version v1.0.0-rc.1
```

Note that the v1 release of the plugin only works with Porter v1.0.0-alpha.20 and higher.

## Plugin Configuration

### Secrets

Secrets plugins allow Porter to store and resolve sensitive bundle data.

For example, to resolve a database password, if your team has a shared key vault that has the password stored in it, you
can use the keyvault plugin to inject it as a credential when you install a bundle.
Another usecase is to store any sensitive bundle parameters and outputs. For
example, if a bundle depends on a redis bundle to generate a database connection
string as an output, the connection string will be securely stored in the key
vault.

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
