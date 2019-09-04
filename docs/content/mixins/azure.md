---
title: azure mixin
description: Using the azure mixin
---

<img src="/images/mixins/azure.png" class="mixin-logo" style="width: 300px"/>

Manage Azure resources

https://github.com/deislabs/porter-azure/

### Install or Upgrade
```
porter mixin install azure
```

### Examples

Create a MySQL database
```yaml
  - azure:
      description: "Create Azure MySQL"
      type: mysql
      name: mysql-azure-porter-demo-wordpress
      resourceGroup: "porter-test"
      parameters:
        administratorLogin: "{{ bundle.parameters.mysql_user }}"
        administratorLoginPassword: "{{ bundle.parameters.mysql_password }}"
        location: "eastus"
        serverName: "{{ bundle.parameters.server_name }}"
        version: "5.7"
        sslEnforcement: "Disabled"
        databaseName: "{{ bundle.parameters.database_name }}"
      outputs:
        - name: "MYSQL_HOST"
          key: "MYSQL_HOST"
```

Create a storage account
```yaml
install:
  - azure:
      description: "Create Azure Storage Account and Container"
      type: storage
      name: porter-azure-ark
      resourceGroup: "{{ bundle.parameters.resource_group }}"
      parameters:
        location: "{{ bundle.parameters.location }}"
        storageAccountName: "{{ bundle.parameters.storage_account_name }}"
        storageContainerName: "{{ bundle.parameters.storage_container_name }}"
      outputs:
        - name: "STORAGE_ACCOUNT_KEY"
          key: "STORAGE_ACCOUNT_KEY"
```
