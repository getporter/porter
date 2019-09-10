---
title: az mixin
description: Using the Azure CLI (az) mixin
---

<img src="/images/mixins/azure.png" class="mixin-logo" style="width: 300px"/>

This is a mixin for Porter that provides the Azure (az) CLI.

Source: https://github.com/deislabs/porter-az

### Install or Upgrade
```
porter mixin install az
```

## Mixin Syntax

See the [az CLI Command Reference](https://docs.microsoft.com/en-us/cli/azure/reference-index?view=azure-cli-latest) for the supported commands.

```yaml
az:
  description: "Description of the command"
  arguments:
  - arg1
  - arg2
  flags:
    FLAGNAME: FLAGVALUE
    REPEATED_FLAG:
    - FLAGVALUE1
    - FLAGVALUE2
  outputs:
    - name: NAME
      jsonPath: JSONPATH
    - name: NAME
      path: SOURCE_FILEPATH
```

### Outputs

The mixin supports `jsonpath` and `path` outputs.


#### JSON Path

The `jsonPath` output treats stdout like a json document and applies the expression, saving the result to the output.

```yaml
outputs:
- name: NAME
  jsonPath: JSONPATH
```

For example, if the `jsonPath` expression was `$[*].id` and the command sent the following to stdout: 

```json
[
  {
    "id": "1085517466897181794",
    "name": "my-vm"
  }
]
```

Then then output would have the following contents:

```json
["1085517466897181794"]
```

#### File Paths

The `path` output saves the content of the specified file path to an output.

```yaml
outputs:
- name: kubeconfig
  path: /root/.kube/config
```

---

## Examples

### Authenticate

```yaml
az:
  description: "Azure CLI login"
  arguments:
    - login
  flags:
    service-principal:
    username: "{{ bundle.credentials.AZURE_SP_CLIENT_ID}}"
    password: "{{ bundle.credentials.AZURE_SP_PASSWORD}}"
    tenant: "{{ bundle.credentials.AZURE_TENANT}}"
```

### Provision a VM

```yaml
az:
  description: "Create VM"
  arguments:
    - vm
    - create
  flags:
    resource-group: porterci
    name: myVM
    image: UbuntuLTS
```

### Delete a VM

```yaml
az:
  description: "Delete VM"
  arguments:
    - vm
    - delete
  flags:
    resource-group: porterci
    name: myVM
```