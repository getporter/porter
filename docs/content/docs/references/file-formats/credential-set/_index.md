---
title: Porter CredentialSet File Format 1.0.1
description: The 1.0.1 file format for Porter CredentialSet resources
layout: single
---

[Credential Sets](/credentials/) can be defined in either json or yaml.
You can use this [json schema][cs-schema] to validate a credential set file.

## Supported Versions

Below are schema versions for credential sets, and the corresponding Porter version that supports it.

| Schema Type   | Schema Version    | Porter Version   |
|---------------|-------------------|------------------|
| CredentialSet | (none)            | v0.38.*          |
| CredentialSet | [1.0.1](./1.0.1/) | v1.0.0-alpha.1+  |

Sometimes you may want to work with a different version of a resource than what is supported by Porter, especially when migrating from one version of Porter to another.
The [schema-check] configuration setting allows you to change how Porter behaves when the schemaVersion of a resource doesn't match Porter's supported version.

[schema-check]: /docs/configuration/configuration/#schema-check

## Example

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

| Field              | Required | Description                                                                                                                   |
|--------------------|----------|-------------------------------------------------------------------------------------------------------------------------------|
| schemaType         | false    | The type of document.                                                                                                         |
| schemaVersion      | true     | The version of the Credential Set schema used in this file.                                                                   |
| name               | true     | The name of the credential set.                                                                                               |
| namespace          | false    | The namespace in which the credential set is defined. Defaults to the empty (global) namespace.                               |
| labels             | false    | A set of key-value pairs associated with the credential set.                                                                  |
| credentials        | true     | A list of credentials and instructions for Porter to resolve the credential value.                                            |
| credentials.name   | true     | The name of the credential as defined in the bundle.                                                                          |
| credentials.source | true     | Specifies how the credential should be resolved. Must have only one child property:<br/> secret, value, env, path, or command |

[cs-schema]: https://raw.githubusercontent.com/getporter/porter/main/pkg/schema/credential-set.schema.json
