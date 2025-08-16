---
title: Porter Plugins File Format 1.0.0
description: The 1.0.0 file format for Porter plugin installation files
layout: single
aliases:
- /reference/file-formats/plugins/1.0.0
---

The plugins file is used when installing multiple plugins at the same time with the [porter plugins install](/cli/porter_plugins_install/) command.
Plugins can be defined in either json or yaml.
You can use this [json schema][plugins-schema] to validate a plugins config file.

## Supported Versions

Below are schema versions for plugin installation files, and the corresponding Porter version that supports it.

| Schema Type | Schema Version | Porter Version |
|-------------|----------------|----------------|
| Plugins     | 1.0.0          | v1.0.6         |

Sometimes you may want to work with a different version of a resource than what is supported by Porter, especially when migrating from one version of Porter to another.
The [schema-check] configuration setting allows you to change how Porter behaves when the schemaVersion of a resource doesn't match Porter's supported version.

[schema-check]: /docs/configuration/configuration/#schema-check

## Changes

```yaml
schemaType: Plugins
schemaVersion: 1.0.0
plugins:
  azure:
    version: v1.0.0
    feedURL: https://cdn.porter.sh/plugins/atom.xml
    url: https://example.com
    mirror: https://example.com
```

| Field                        | Required | Description                                                 |
|------------------------------|----------|-------------------------------------------------------------|
| schemaType                   | false    | The type of document.                                       |
| schemaVersion                | true     | The version of the Plugins schema used in this file.        |
| plugins.<pluginName>.version | false    | The version of the plugin.                                  |
| plugins.<pluginName>.feedURL | false    | The url of an atom feed where the plugin can be downloaded. |
| plugins.<pluginName>.url     | false    | The url from where the plugin can be downloaded.            |
| plugins.<pluginName>.mirror  | false    | The mirror of official Porter assets.                       |


[plugins-schema]: https://raw.githubusercontent.com/getporter/porter/main/pkg/schema/plugins.schema.json