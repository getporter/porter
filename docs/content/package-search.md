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

Porter maintains lists for mixins and plugins available for installation.
They are represented in JSON and hosted in the [porter-packages][porter-packages]
GitHub repository:

* [Mixin Directory](https://github.com/getporter/packages/blob/master/mixins/index.json)
* [Plugin Directory](https://github.com/getporter/packages/blob/master/plugins/index.json)

To add a listing for your mixin or plugin for others to see, follow the
instructions in the [repository README][porter-packages-readme].

[porter-packages]: https://github.com/getporter/packages
[porter-packages-readme]: https://github.com/getporter/packages/blob/master/README.md