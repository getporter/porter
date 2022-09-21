---
title: az mixin
description: Run Azure commands using the az CLI
---

<img src="/images/mixins/azure.png" class="mixin-logo" style="width: 300px"/>

Run Azure commands using the az CLI.

Source: https://github.com/getporter/az-mixin

### Install or Upgrade
```
porter mixin install az --version v1.0.0-rc.1
```

## Mixin Configuration

When you declare the mixin, you can also configure additional extensions to install:

**Use the vanilla az CLI**
```yaml
mixins:
- az
```

**Install additional extensions**

```yaml
mixins:
- az:
    extensions:
    - EXTENSION_NAME
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
  supress-output: false
  outputs:
    - name: NAME
      jsonPath: JSONPATH
    - name: NAME
      path: SOURCE_FILEPATH
```

NOTE: Some commands may not allow a flag to be repeated, and use a different
syntax such as packing all the values into a single flag instance. [Change
Settings for a Web Application](#change-settings-for-a-web-application)
demonstrates how to handle inconsistent flags behavior.

### Suppress Output

The `suppress-output` field controls whether output from the mixin should be
prevented from printing to the console. By default this value is false, using
Porter's default behavior of hiding known sensitive values. When 
`suppress-output: true` all output from the mixin (stderr and stdout) are hidden.

Step outputs (below) are still collected when output is suppressed. This allows
you to prevent sensitive data from being exposed while still collecting it from
a command and using it in your bundle.

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
  path: /home/nonroot/.kube/config
```

---

## Examples

### Install the Azure IoT Extension

```yaml
mixins:
- az:
    extensions:
    - azure-cli-iot-ext
```

### Authenticate

```yaml
az:
  description: "Azure CLI login"
  arguments:
    - login
  flags:
    service-principal:
    username: ${ bundle.credentials.AZURE_SP_CLIENT_ID}
    password: ${ bundle.credentials.AZURE_SP_PASSWORD}
    tenant: ${ bundle.credentials.AZURE_TENANT}
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

### Change Settings for a Web Application

The `--settings` flag for this command does not support being repeated. Instead you must pack all
the setting values into a single flag using space-separated KEY=VALUE pairs.

```yaml
install: 
  - az:
      description: 'Deploy Web API configurations'
      arguments:
        - webapp
        - config
        - appsettings
        - set
      flags:
        ids: '${ bundle.outputs.WEBAPI_ID }'
        settings: 'PGHOST=${ bundle.outputs.POSTGRES_HOST } PGUSER=${ bundle.outputs.POSTGRES_USER } PGPASSWORD=${ bundle.outputs.POSTGRES_PASSWORD } PGDB=${ bundle.outputs.POSTGRES_DB }'
```
