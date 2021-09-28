---
title: File Formats
---

* [Credential Sets](#credential-set)
* [Parameter Sets](#parameter-set)
* [Installation](#installation)

## Credential Set

Credential sets can be defined in either json or yaml.
You can use this [json schema][cs-schema] to validate a credential set file.

```yaml
schemaVersion: 1.0.0
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

| Field  | Required  | Description  |
|---|---|---|
| schemaVersion  | true  | The version of the Credential Set schema used in this file.  |
| name  | true  | The name of the credential set.  |
| namespace  | false  | The namespace in which the credential set is defined. Defaults to the empty (global) namespace.  |
| labels  | false | A set of key-value pairs associated with the credential set. |
| credentials | true | A list of credentials and instructions for Porter to resolve the credential value. |
| credentials.name | true | The name of the credential as defined in the bundle. |
| credentials.source | true | Specifies how the credential should be resolved. Must have only one child property:<br/> secret, value, env, path, or command |

## Parameter Set

Parameter sets can be defined in either json or yaml.
You can use this [json schema][ps-schema] to validate a parameter set file.

```yaml
schemaVersion: 1.0.0
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

| Field  | Required  | Description  |
|---|---|---|
| schemaVersion  | true  | The version of the Parameter Set schema used in this file.  |
| name  | true  | The name of the parameter set.  |
| namespace  | false  | The namespace in which the parameter set is defined. Defaults to the empty (global) namespace.  |
| labels  | false | A set of key-value pairs associated with the parameter set. |
| parameters | true | A list of parameters and instructions for Porter to resolve the parameter value. |
| parameters.name | true | The name of the parameter as defined in the bundle. |
| parameters.source | true | Specifies how the parameter should be resolved. Must have only one child property:<br/> secret, value, env, path, or command |

## Installation

Installations can be defined in either json or yaml.
You can use this [json schema][inst-schema] to validate an installation file.

Either bundleVersion, bundleTag or bundleDigest must be specified.
When the bundleDigest is specified in addition to the version or tag, it is used to validate the bundle that was pulled using the other fields.

```yaml
schemaVersion: 1.0.0
name: myinstallation
namespace: staging
labels:
  team: marketing
  customer: bigbucks
bundleRepository: getporter/porter-hello
# Only one of the following fields must be specified: bundleVersion, bundleDigest, or bundleTag
bundleVersion: 0.1.1
bundleDigest: sha256:ace0eda3e3be35a979cec764a3321b4c7d0b9e4bb3094d20d3ff6782961a8d54
bundleTag: latest
parameterSets:
  - myparams
credentialSets:
  - mycreds
parameters:
  log-level: 11
```

| Field  | Required  | Description  |
|---|---|---|
| schemaVersion  | true  | The version of the Installation schema used in this file.  |
| name  | true  | The name of the parameter set.  |
| namespace  | false  | The namespace in which the parameter set is defined. Defaults to the empty (global) namespace.  |
| labels  | false | A set of key-value pairs associated with the parameter set. |
| bundleRepository | true | The repository where the bundle is published. | 
| bundleVersion | false* | The bundle version. |
| bundleTag | false* | The bundle tag. This is useful when you do not use Porter's convention of defaulting the bundle tag to the bundle version. |
| bundleDigest | false* | The bundle repository digest. |
| parameterSets | false | A list of parameter set names. |
| credentialSets | false | A list of credential set names. |
| parameters | false | Additional parameter values to use with the installation. Overrides any parameters defined in the associated parameter sets. |


[cs-schema]: https://raw.githubusercontent.com/getporter/porter/release/v1/pkg/schema/credential-set.schema.json
[ps-schema]: https://raw.githubusercontent.com/getporter/porter/release/v1/pkg/schema/parameter-set.schema.json
[inst-schema]: https://raw.githubusercontent.com/getporter/porter/release/v1/pkg/schema/installation.schema.json