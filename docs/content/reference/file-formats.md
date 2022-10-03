---
title: File Formats
description: Defines the format of files used by Porter
---

* [Supported Versions](#supported-versions)
* [Manifest](/bundle/manifest/file-format/)
* [Credential Sets](#credential-set)
* [Parameter Sets](#parameter-set)
* [Installation](#installation)
* [Porter Operator File Formats](/operator/file-formats/)

## Supported Versions

Below are schema versions for each of the file formats, and the corresponding Porter version that supports it.

| Schema Type   | Schema Version                                               | Porter Version   |
|---------------|--------------------------------------------------------------|------------------|
| Bundle        | (none)                                                       | v0.38.*          |
| Bundle        | [1.0.0-alpha.1](/bundle/manifest/file-format/1.0.0-alpha.1/) | v1.0.0-alpha.14+ |
| Bundle        | [1.0.0](/bundle/manifest/file-format/1.0.0/)                 | v1.0.0-beta.2+   |
| CredentialSet | (none)                                                       | v0.38.*          |
| CredentialSet | 1.0.1                                                        | v1.0.0-alpha.1+  |
| ParameterSet  | (none)                                                       | v0.38.*          |
| ParameterSet  | 1.0.1                                                        | v1.0.0-alpha.1+  |
| Installation  | 1.0.1                                                        | v1.0.0-alpha.20+ |

Sometimes you may want to work with a different version of a resource than what is supported by Porter, especially when migrating from one version of Porter to another.
The [schema-check] configuration setting allows you to change how Porter behaves when the schemaVersion of a resource doesn't match Porter's supported version.

[schema-check]: /configuration/#schema-check

## Credential Set

Credential sets can be defined in either json or yaml.
You can use this [json schema][cs-schema] to validate a credential set file.

```yaml
schemaType: CredentialSet
schemaVersion: 1.0.1
name: mycreds
namespace: staging
labels:
  team: redteam
  owner: xianglu
credentials:
  - name: token
    source:
      env: GITHUB_TOKEN
  - name: kubeconfig
    source:
      path: $HOME/.kube/config
  - name: connStr
    source:
      secret: my-connection-string
```

| Field              | Required | Description                                                                                                                                    |
|--------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------|
| schemaType         | false    | The type of document. This isn't used by Porter but is included when Porter outputs the file, so that editors can determine the resource type. |
| schemaVersion      | true     | The version of the Credential Set schema used in this file.                                                                                    |
| name               | true     | The name of the credential set.                                                                                                                |
| namespace          | false    | The namespace in which the credential set is defined. Defaults to the empty (global) namespace.                                                |
| labels             | false    | A set of key-value pairs associated with the credential set.                                                                                   |
| credentials        | true     | A list of credentials and instructions for Porter to resolve the credential value.                                                             |
| credentials.name   | true     | The name of the credential as defined in the bundle.                                                                                           |
| credentials.source | true     | Specifies how the credential should be resolved. Must have only one child property:<br/> secret, value, env, path, or command                  |

## Parameter Set

Parameter sets can be defined in either json or yaml.
You can use this [json schema][ps-schema] to validate a parameter set file.

```yaml
schemaType: ParameterSet
schemaVersion: 1.0.1
name: myparams
namespace: staging
labels:
  team: redteam
  owner: xianglu
parameters:
  - name: color
    source:
      value: red
  - name: log-level
    source:
      env: LOG_LEVEL
  - name: connStr
    source:
      secret: my-connection-string
```

| Field             | Required | Description                                                                                                                                    |
|-------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------|
| schemaType        | false    | The type of document. This isn't used by Porter but is included when Porter outputs the file, so that editors can determine the resource type. |
| schemaVersion     | true     | The version of the Parameter Set schema used in this file.                                                                                     |
| name              | true     | The name of the parameter set.                                                                                                                 |
| namespace         | false    | The namespace in which the parameter set is defined. Defaults to the empty (global) namespace.                                                 |
| labels            | false    | A set of key-value pairs associated with the parameter set.                                                                                    |
| parameters        | true     | A list of parameters and instructions for Porter to resolve the parameter value.                                                               |
| parameters.name   | true     | The name of the parameter as defined in the bundle.                                                                                            |
| parameters.source | true     | Specifies how the parameter should be resolved. Must have only one child property:<br/> secret, value, env, path, or command                   |

## Installation

Installations can be defined in either json or yaml.
You can use this [json schema][inst-schema] to validate an installation file.

Either the bundle digest, version, or tag must be specified.
When more than one is specified, Porter selects the most specific field available, preferring digest the most, then version, and then falling back to tag last.

```yaml
schemaType: Installation
schemaVersion: 1.0.0
name: myinstallation
namespace: staging
uninstalled: false
labels:
  team: marketing
  customer: bigbucks
bundle:
  repository: ghcr.io/getporter/examples/porter-hello
  # One of the following fields must be specified: digest, version, or tag
  digest: sha256:276b44be3f478b4c8d1f99c1925386d45a878a853f22436ece5589f32e9df384
  version: 0.2.0
  tag: latest
parameterSets:
  - myparams
credentialSets:
  - mycreds
parameters:
  log-level: 11
```

| Field             | Required | Description                                                                                                                                    |
|-------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------|
| schemaType        | false    | The type of document. This isn't used by Porter but is included when Porter outputs the file, so that editors can determine the resource type. |
| schemaVersion     | true     | The version of the Installation schema used in this file.                                                                                      |
| name              | true     | The name of the installation.                                                                                                                  |
| namespace         | false    | The namespace in which the installation is defined. Defaults to the empty (global) namespace.                                                  |
| uninstalled       | false    | Specifies if the installation should be uninstalled. Defaults to false.                                                                        |
| labels            | false    | A set of key-value pairs associated with the installation.                                                                                     |
| bundle            | true     | A reference to where the bundle is published                                                                                                   |
| bundle.repository | true     | The repository where the bundle is published.                                                                                                  | 
| bundle.digest     | false*   | The bundle repository digest.                                                                                                                  |
| bundle.version    | false*   | The bundle version.                                                                                                                            |
| bundle.tag        | false*   | The bundle tag. This is useful when you do not use Porter's convention of defaulting the bundle tag to the bundle version.                     |
| parameterSets     | false    | A list of parameter set names.                                                                                                                 |
| credentialSets    | false    | A list of credential set names.                                                                                                                |
| parameters        | false    | Additional parameter values to use with the installation. Overrides any parameters defined in the associated parameter sets.                   |

\* The bundle section requires a repository and one of the following fields: digest, version, or tag.

[cs-schema]: /schema/v1/credential-set.schema.json
[ps-schema]: /schema/v1/parameter-set.schema.json
[inst-schema]: /schema/v1/installation.schema.json
