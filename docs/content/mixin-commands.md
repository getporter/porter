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

* [version](#version)
* [status](#status)

# build

The build command (required) is called on the local machine during the `porter
build` command. The mixin should return lines for the Dockerfile on stdout.

Example:

```console
$ helm build
RUN apt-get update && \
    apt-get install -y curl && \
    curl -o helm.tgz https://storage.googleapis.com/kubernetes-helm/helm-v2.12.3-linux-amd64.tar.gz && \
    tar -xzf helm.tgz && \
    mv linux-amd64/helm /usr/local/bin && \
    rm helm.tgz
RUN helm init --client-only
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

**NOTE**: porter handles rewriting the references when it merges the json
*schemas. So write your references relative to your mixin's schema document, and
*porter will take care of adjusting it when the schema is merged.

# install

The install command (required) is called from inside the invocation image during
the `porter run` command. The current step from the manifest is passed on stdin.
The mixin should write any output values to a file in the
`/cnab/app/porter/outputs/` directory.

Example:

**stdin**
```yaml
install:
- helm:
    description: "Install MySQL"
    name: porter-ci-mysql
    chart: stable/mysql
```

**/cnab/app/porter/outputs/mixin-output-abc123**
```
password=topsecret
host=127.0.0.1
```

# upgrade

The upgrade command (required) is called from inside the invocation image during
the `porter run` command. The current step from the manifest is passed on stdin.
The mixin should write any output values to a file in the
`/cnab/app/porter/outputs/` directory.

Example:

**stdin**
```yaml
upgrade:
- helm:
    description: "Upgrade MySQL"
    name: porter-ci-mysql
```

**/cnab/app/porter/outputs/mixin-output-abc123**
```
password=topsecret
host=127.0.0.1
```

# uninstall

The upgrade command (required) is called from inside the invocation image during
the `porter run` command. The current step from the manifest is passed on stdin.
The mixin should write any output values to a file in the
`/cnab/app/porter/outputs/` directory.

Example:

**stdin**
```yaml
upgrade:
- helm:
    description: "Upgrade MySQL"
    name: porter-ci-mysql
```

**/cnab/app/porter/outputs/mixin-output-abc123**
```
password=topsecret
host=127.0.0.1
```

# status

The status command (required) is called from inside the invocation image during
the `porter run` command. The current step from the manifest is passed on stdin.
The mixin should write any output values to a file in the
`/cnab/app/porter/outputs/` directory.

Example:

**stdin**
```yaml
status:
- helm:
    description: "Status of MySQL"
    releases:
    - porter-ci-mysql

# version

The version command (optional) is not used by porter. We encourage you to
implement it to be consistent with the other mixins and help users know that
they have the correct version of your mixin installed.
 
Example:

```console
$ ~/.porter/mixins/exec/exec version
exec mixin v0.4.0-ralpha.1+dubonnet (2aa921d)
```



