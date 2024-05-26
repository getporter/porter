---
title: Porter ParameterSet File Format 1.0.1
description: The 1.0.1 file format for Porter ParameterSet resources
layout: single
---

[Parameter sets](/parameters/) can be defined in either json or yaml.
You can use this [json schema][ps-schema] to validate a parameter set file.

## Supported Versions

Below are schema versions for parameter sets, and the corresponding Porter version that supports it.

| Schema Type  | Schema Version    | Porter Version  |
|--------------|-------------------|-----------------|
| ParameterSet | (none)            | v0.38.*         |
| ParameterSet | [1.0.1](./1.0.1/) | v1.0.0-alpha.1+ |

Sometimes you may want to work with a different version of a resource than what is supported by Porter, especially when migrating from one version of Porter to another.
The [schema-check] configuration setting allows you to change how Porter behaves when the schemaVersion of a resource doesn't match Porter's supported version.

[schema-check]: /docs/configuration/configuration/#schema-check

## Example

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

| Field             | Required | Description                                                                                                                  |
|-------------------|----------|------------------------------------------------------------------------------------------------------------------------------|
| schemaType        | false    | The type of document.                                                                                                        |
| schemaVersion     | true     | The version of the Parameter Set schema used in this file.                                                                   |
| name              | true     | The name of the parameter set.                                                                                               |
| namespace         | false    | The namespace in which the parameter set is defined. Defaults to the empty (global) namespace.                               |
| labels            | false    | A set of key-value pairs associated with the parameter set.                                                                  |
| parameters        | true     | A list of parameters and instructions for Porter to resolve the parameter value.                                             |
| parameters.name   | true     | The name of the parameter as defined in the bundle.                                                                          |
| parameters.source | true     | Specifies how the parameter should be resolved. Must have only one child property:<br/> secret, value, env, path, or command |

[ps-schema]: https://raw.githubusercontent.com/getporter/porter/main/pkg/schema/parameter-set.schema.json
