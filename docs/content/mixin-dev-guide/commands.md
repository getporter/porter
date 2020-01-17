---
title: Mixin Commands
description: How to implement required and optional commands for your custom Porter mixin
---

These are the commands that a mixin can implement to work with porter. Some must
be implemented, or porter will refuse to load the mixin, while others are just
recommended so that your mixin has a good user experience.

**Required Commands**

* [build](#build)
* [schema](#schema)
* [install](#install)
* [upgrade](#upgrade)
* [uninstall](#uninstall)

**Optional Commands**

* [invoke](#invoke)
* [version](#version)


# build

The build command (required) is called on the local machine during the `porter
build` command. Any mixin configuration and all usages of the mixin are passed
on stdin. The mixin should return lines for the Dockerfile on stdout.

Example:

**stdin**

```yaml
config:
  extensions:
  - iot
actions:
  install:
  - az:
      arguments:
      - login
      description: Login
  uninstall: []
  upgrade: []
```

**stdout**
```console
RUN apt-get update && apt-get install -y apt-transport-https lsb-release gnupg curl
RUN curl -sL https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > /etc/apt/trusted.gpg.d/microsoft.asc.gpg
RUN echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ $(lsb_release -cs) main" > /etc/apt/sources.list.d/azure-cli.list
RUN apt-get update && apt-get install -y azure-cli
RUN az extension add --name azure-cli-iot-ext 
```

# schema

The schema command (required) is used in multiple porter commands, such as
`porter schema`, `porter build` and `porter run`. The mixin should return a json
schema for the parts of the Porter manifest that it is responsible for. Porter
combines each of the mixin's schema documents into a single json schema document
and uses that to validate the Porter manifest, and other tools, like VS Code,
use the document to provide autocomplete when editing the manifest.

There are a few rules to follow when authoring your mixin's json schema as there
are known bugs in some of the libraries in the libraries used by Porter and VS
Code that they workaround:

1. Do not chain together references. A reference should only link to another type,
not immediately to another `$ref`.

1. Only use relative references, such as `#/definitions/installStep`.

1. Do not use `$id`'s. Do not use references to an internal `$id`.

1. Every supported action must be defined with a root level property named after
the action, e.g. "install", and a definition named `<action>Step`, e.g.
"installStep". Below is an example from the exec mixin for the install action:

    ```json
    {
      "definitions": {
        "installStep": {
          "type": "object",
          "properties": {
            "exec": {
              "type": "object",
              "properties": {
                "description": {
                  "type": "string",
                  "minLength": 1
                },
                "command": {
                  "type": "string"
                },
                "arguments": {
                  "type": "array",
                  "items": {
                    "type": "string",
                    "minItems": 1
                  }
                }
              },
              "additionalProperties": false,
              "required": [
                "description",
                "command"
              ]
            }
          },
          "additionalProperties": false,
          "required": [
            "exec"
          ]
        }
      },
      "type": "object",
      "properties": {
        "install": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/installStep"
          }
        }
      },
      "additionalProperties": false
    }
    ```
1. Custom action support is signaled by having a property named ".*" with items
   of type `invokeStep`.

**NOTE**: porter handles rewriting the references when it merges the json
*schemas. So write your references relative to your mixin's schema document, and
*porter will take care of adjusting it when the schema is merged.


# install

The install command (required) is called from inside the invocation image during
the `porter run` command. The current step from the manifest is passed on stdin.
The mixin should write any output values to their own files in the
`/cnab/app/porter/outputs/` directory.

Example:

**stdin**
```yaml
install:
- helm:
    description: "Install MySQL"
    name: porter-ci-mysql
    chart: stable/mysql
    outputs:
      - name: mysql-root-password
        secret: "{{ bundle.parameters.mysql-name }}"
        key: mysql-root-password
      - name: mysql-password
        secret: "{{ bundle.parameters.mysql-name }}"
        key: mysql-password
```

**/cnab/app/porter/outputs/mysql-root-password**
```
topsecret
```
**/cnab/app/porter/outputs/mysql-password**
```
alsotopsecret
```

# upgrade

The upgrade command (required) is called from inside the invocation image during
the `porter run` command. The current step from the manifest is passed on stdin.
The mixin should write any output values to their own files in the
`/cnab/app/porter/outputs/` directory.

Example:

**stdin**
```yaml
parameters:
  - name: mysql-password
    type: string
...
upgrade:
- helm:
    description: "Upgrade MySQL"
    name: porter-ci-mysql
    replace: true
    set:
      mysqlPassword: "{{ bundle.parameters.mysql-password }}"
```

**/cnab/app/porter/outputs/mysql-root-password**
```
topsecret
```
**/cnab/app/porter/outputs/mysql-password**
```
updatedtopsecret
```

# uninstall

The uninstall command (required) is called from inside the invocation image during
the `porter run` command. The current step from the manifest is passed on stdin.

Example:

**stdin**
```yaml
uninstall:
- helm:
    description: "Uninstall MySQL"
    purge: true
    releases:
      - "{{ bundle.parameters.mysql-name }}"
```

# invoke

The invoke command (optional) is called from inside the invocation image during
the `porter run` command when a custom action defined in the bundle is executed.
The current step from the manifest is passed on stdin. The mixin should write
any output values to their own files in the `/cnab/app/porter/outputs/` directory.

A mixin doesn't have to support custom actions. If your mixin just maps to CLI
commands, and doesn't take the action into account, then we strongly recommend
supporting custom actions so that bundle authors can use your mixin anywhere in
their bundle, not just in install/upgrade/uninstall.

Example:

**stdin**
```yaml
status:
- exec:
    description: "Run a rando command, don't care what action I'm in"
    command: bash
    flags:
      c: echo "Don't mind me, just getting the status of something..."
```

# version

The version command (optional) is used by porter when listing installed mixins 
via `porter mixins list`. It should support an `--output|o` flag that accepts
either `plaintext` or `json` as values, defaulting to `plaintext`.
 
Example:

```console
$ ~/.porter/mixins/exec/exec version
exec mixin v0.13.1-beta.1 (37f3637)

$ ~/.porter/mixins/exec/exec version --output json
{
  "name": "exec",
  "version": "v0.13.1-beta.1",
  "commit": "37f3637",
  "author": "Porter Authors"
}
```
