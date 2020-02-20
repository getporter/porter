---
title: Mixin and Plugin Search
description: Find a plugin for Porter and list your own plugin in our search results
---

## Search

Porter maintains lists for mixins and plugins available for users to install.
These can be searched via [porter mixin search](/cli/porter_mixin_search/) or
[porter plugin search](/cli/porter_plugin_search/).

For example, here we search for mixins with the term `az` in the name:

```console
$ porter mixin search az
Name   Description                    Author           URL                                     URL Type
az     A mixin for using the az cli   Porter Authors   https://cdn.porter.sh/mixins/atom.xml   Atom Feed
```

If no query is supplied, the full listing will be returned.

For example, here we search for all plugins, specifying `yaml` output:

```console
$ porter plugin search -o yaml
- name: azure
  author: Porter Authors
  description: Integrate Porter with Azure. Store Porter's data in Azure Cloud and
    secure your bundle's secrets in Azure Key Vault.
  url: https://cdn.porter.sh/plugins/atom.xml
...
```

## List

Porter maintains a list each for mixins and plugins available for installation.
They are represented in JSON:

* [Mixin Directory](https://github.com/deislabs/porter/blob/master/pkg/mixins/directory/index.json)
* [Plugin Directory](https://github.com/deislabs/porter/blob/master/pkg/plugins/directory/index.json)

To list your mixin or plugin for others to see, create a new JSON entry just
like the others, with details updated to reflect your offering.

For example, a new plugin entry would look like:

```json
  {
    "name": "myplugin",
    "author": "My Name",
    "description": "A plugin for doing great things",
    "URL": "https://github.com/org/project/releases/download",
  },
```

The `URL` field should be one of the following:

* **Atom Feed URL:** Porter uses the following for its stable plugins: `https://cdn.porter.sh/plugin/atom.xml`
* **Download URL:** Directory where binaries are hosted, such as GitHub releases: `https://github.com/org/project/releases/download`

Open up a pull request with the updated directory.  Once merged, your mixin or
plugin listing will be broadcast to the world!

