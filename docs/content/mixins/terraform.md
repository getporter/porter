---
title: terraform mixin
description: Using the terraform mixin
---

Execute Terraform files

Source: https://github.com/deislabs/porter-terraform

### Install or Upgrade
```
porter mixin install terraform --feed-url https://cdn.deislabs.io/porter/atom.xml
```

### Examples

```yaml
install:
  - terraform:
      description: "Install Azure Key Vault"
      autoApprove: true
      input: false
      backendConfig:
        key: "{{ bundle.name }}.tfstate"
        storage_account_name: "{{ bundle.credentials.backend_storage_account }}"
        container_name: "{{ bundle.credentials.backend_storage_container }}"
        access_key: "{{ bundle.credentials.backend_storage_access_key }}"
      outputs:
- name: vault_uri
```
