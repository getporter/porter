---
title: Plugin Discoverability
description: Learn more about searching and broadcasting available plugins for Porter
---

## Search

Porter maintains a list of plugins that users can search via
`porter plugin search [NAME]`. If no name is supplied, the full listing will be
returned.  See the help menu for all command options: `porter plugin search -h`.

For example, here we search for an `azure` plugin:

```console
$ porter plugin search azure
Name    Description                                    Author           URL                                      URL Type
azure   A plugin for utilizing Azure cloud resources   Porter Authors   https://cdn.porter.sh/plugins/atom.xml   Atom Feed
```

## Broadcast

To add your plugin to the list, add an entry to the plugin directory
[list](https://github.com/deislabs/porter/blob/master/pkg/plugins/directory/index.json)
with all the pertinent informational fields filled out.

For instance, a new entry might look like:

```json
  {
    "name": "myplugin",
    "author": "My Name",
    "description": "A plugin for doing great things",
    "URL": "https://github.com/org/project/releases/download",
  },
```

The `URL` field should either be an Atom Feed URL (for example, Porter uses
the following for its stable plugins: `https://cdn.porter.sh/plugin/atom.xml`) or
a download URL (like the GitHub download URL shown above:
`https://github.com/org/project/releases/download`)

With this change pushed to a branch on your fork of this repo, you're
now ready to open up a [Pull Request](https://github.com/deislabs/porter/pulls).
After the changes are approved, merged and included in the next release, your
plugin will start to show up in the official listing!
