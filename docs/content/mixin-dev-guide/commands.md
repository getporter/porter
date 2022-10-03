---
title: Mixin Commands
description: How to implement required and optional commands for your custom Porter mixin
---

These are the commands that a mixin can implement to work with porter. Some must
be implemented, or porter will refuse to load the mixin, while others are just
recommended so that your mixin has a good user experience.

Our [skeleton mixin template][skeletor] demonstrates how to implement each command,
providing a working implementation, tests, and a Makefile to manage common tasks. If
you are writing a mixin in Go, we strongly recommend starting from the template.

**Required Commands**

* [build](#build)
* [schema](#schema)
* [install](#install)
* [upgrade](#upgrade)
* [uninstall](#uninstall)
* [version](#version)

**Optional Commands**

* [invoke](#invoke)
* [lint](#lint)


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
schema for the parts of the Porter manifest it is responsible for. Porter
combines each of the mixin's schema documents into a single json schema document
and uses that to validate the Porter manifest, and other tools, like VS Code,
use the document to provide autocomplete when editing the manifest.

The output of your schema command must be a [json schema][jsonschema] document.
The schema should describe which fields are supported by your mixin, and their
descriptions, type and optionality. Good examples of a mixin schema are the
[exec mixin schema] and [helm mixin schema].

Your mixin is responsible for defining the allowed schema for each action
because for some mixins the schema is different per action. Porter will handle
parsing through your schema and piecing it all together into a coherent schema
for porter.yaml. You just need to worry about what your schema should look like
for your portion of the manifest related to the mixin:

Below is an example of the potion of the Porter manifest that will be used with the 
[exec mixin schema].

```yaml
install:
- exec:
   description: Some description
   command: ./helpers.sh
```

If your mixin supports additional configuration when it is declared in porter.yaml,
you should define that in your schema in the "config" definition. 

Here is an example of how the Kubernetes mixin schema defines its configuration schema:

```yaml
mixins:
- kubernetes:
   clientVersion: 1.2.3
```

Below is a partial json schema for the Kubernetes mixin that only shows the config section.
The config section should contain a single property named after the mixin, that contains the mixin's configuration schema.

```json
{
  "definitions": {
    "config": {
      "description": "Configuration that can be set when the mixin is declared",
      "type": "object",
      "properties": {
        "kubernetes": {
          "description": "kubernetes mixin configuration",
          "type": "object",
          "properties": {
            "clientVersion": {
              "description": "Version of kubectl to install in the bundle",
              "type": "string"
            }
          },
          "additionalProperties": false
        }
      },
      "additionalProperties": false
    }
  }
}
```


The [mixin skeleton template][skeletor] provides an example implementation, unit tests
and an integration test to validate your implementation. After you have customized
your schema command, you can test it out with the [Porter extension](https://marketplace.visualstudio.com/items?itemName=ms-kubernetes-tools.porter-vscode)
for Visual Studio Code. Install your updated mixin, and then open a porter.yaml
file with VS Code. You should get autocomplete and hover documentation for your
mixin's fields.

A great resource for developing and testing your schema is the [JSON Schema Validator]
and [YAML to JSON converter].

There are a few rules to follow when authoring your mixin's json schema as there
are known bugs in the libraries used by Porter and VS
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
- helm3:
    description: "Install MySQL"
    name: porter-ci-mysql
    chart: bitnami/mysql
    outputs:
      - name: mysql-root-password
        secret: ${ bundle.parameters.mysql-name }
        key: mysql-root-password
      - name: mysql-password
        secret: ${ bundle.parameters.mysql-name }
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
- helm3:
    description: "Upgrade MySQL"
    name: porter-ci-mysql
    replace: true
    set:
      mysqlPassword: ${ bundle.parameters.mysql-password }
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
- helm3:
    description: "Uninstall MySQL"
    purge: true
    releases:
      - ${ bundle.parameters.mysql-name }
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

The version command (required) is used by porter during `porter build` and when
listing installed mixins via `porter mixins list`. It should support an
`--output|o` flag that accepts either `plaintext` or `json` as values,
defaulting to `plaintext`.
 
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

[jsonschema]: https://json-schema.org/understanding-json-schema/
[skeletor]: https://github.com/getporter/skeletor
[JSON Schema Validator]: https://www.jsonschemavalidator.net/
[YAML to JSON converter]: https://www.convertjson.com/yaml-to-json.htm
[exec mixin schema]: /src/pkg/exec/schema/exec.json
[helm mixin schema]: /helm-mixin/src/pkg/helm/schema/schema.json