---
title: terraform mixin
description: Execute Terraform files using the terraform CLI
---

<img src="/images/mixins/terraform.svg" class="mixin-logo" style="width: 300px" />

Execute Terraform files using the [terraform CLI](https://www.terraform.io/)

Source: https://github.com/getporter/terraform-mixin

### Install or Upgrade
```
porter mixin install terraform --version v1.0.0-rc.1
```

### Examples

```yaml
install:
  - terraform:
      description: "Install Azure Key Vault"
      input: false
      backendConfig:
        key: ${ bundle.name }.tfstate"
        storage_account_name: ${ bundle.credentials.backend_storage_account }
        container_name: ${ bundle.credentials.backend_storage_container }
        access_key: ${ bundle.credentials.backend_storage_access_key }
      outputs:
- name: vault_uri
```
