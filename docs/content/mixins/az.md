---
title: az mixin
description: Run Azure commands using the az CLI
---

<img src="/images/mixins/azure.png" class="mixin-logo" style="width: 300px"/>

Run Azure commands using the az CLI.

Source: https://github.com/getporter/az-mixin

### Install or Upgrade
```
porter mixin install az
```

## Mixin Configuration

### Client Version
By default, the most recent version of the az CLI is installed.
You can specify a specific version with the `clientVersion` setting.

```yaml
mixins:
  - az:
      clientVersion: 1.2.3
```

### Extensions

When you declare the mixin, you can also configure additional extensions to install

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

### User Agent Opt Out

When you declare the mixin, you can disable the mixin from customizing the az user agent string

```yaml
mixins:
- az:
    userAgentOptOut: true
```

By default, the az mixin adds the porter and mixin version to the user agent string used by the az CLI.
We use this to understand which version of porter and the mixin are being used by a bundle, and assist with troubleshooting.
Below is an example of what the user agent string looks like:

```
AZURE_HTTP_USER_AGENT="getporter/porter/v1.0.0 getporter/az/v1.2.3"
```

You can add your own custom strings to the user agent string by editing your [template Dockerfile] and setting the AZURE_HTTP_USER_AGENT environment variable.

[template Dockerfile]: https://getporter.org/bundle/custom-dockerfile/

## Mixin Syntax

The format below is for executing any arbitrary az CLI command.

See the [az CLI Command Reference](https://docs.microsoft.com/en-us/cli/azure/reference-index?view=azure-cli-latest) for the supported commands.

```yaml
az:
  description: "Description of the command"
  arguments: # arguments to pass to the az CLI
  - arg1
  - arg2
  flags: # flags to pass to the az CLI, porter determines if it is a long (--flag) or short flag (-f)
    a: flag-value
    long-flag: true
    repeated-flag:
    - flag-value1
    - flag-value2
  suppress-output: false
  ignoreError: # Conditions when execution should continue even if the command fails
    all: true # Ignore all errors
    exitCodes: # Ignore failed commands that return the following exit codes
      - 1
      - 2
    output: # Ignore failed commands based on the contents of stderr
      contains: # Ignore when stderr contains a substring
        - "SUBSTRING IN STDERR"
      regex: # Ignore when stderr matches a regular expression
        - "GOLANG_REGULAR_EXPRESSION"
  outputs: # Collect values from the command and make it available as an output
    - name: NAME
      jsonPath: JSONPATH # Scrape stdout with a json path expression
    - name: NAME
      regex: GOLANG_REGULAR_EXPRESSION # Scrape stdout with a regular expression
    - name: NAME
      path: FILEPATH # Save the contents of a file
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

### Ignore Error

In some cases, you may need to have the bundle continue executing when a command fails.
For example when the command fails because the resource already exists.

You can ignore errors based on:

* All - Ignore all errors from the command.
* ExitCodes - Ignore errors when one of the specified exit codes are returned.
* Output Contains - Ignore errors when the command's stderr contains the specified string.
* Output Regex - Ignore errors when the command's stderr matches the specified regular expression (in Go syntax).

Porter only prints out that an error was ignored in debug mode.

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

## Resource Groups

The mixin has a custom command to manage a resource group.

When used in any action other than uninstall, the mixin will ensure that the resource group exists.
```yaml
install:
  - az:
      description: "Ensure my group exists"
      group:
        name: mygroup
        location: westus
```

When used in the uninstall action, the mixin ensures that the resource group is deleted.
```yaml
uninstall:
  - az:
      description: "Cleanup my group"
      group:
        name: mygroup
```

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
    username: ${ bundle.credentials.AZURE_SP_CLIENT_ID }
    password: ${ bundle.credentials.AZURE_SP_PASSWORD }
    tenant: ${ bundle.credentials.AZURE_TENANT }
```

### Provision a VM

Create a VM, ignoring the error if it already exists.

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
  ignoreErrors:
    output:
      contains: ["already exists"]
```

### Delete a VM

Delete a VM, ignoring the error if it has already been removed.

```yaml
az:
  description: "Delete VM"
  arguments:
    - vm
    - delete
  flags:
    resource-group: porterci
    name: myVM
  ignoreErrors:
    output:
      contains: ["not found"]
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
