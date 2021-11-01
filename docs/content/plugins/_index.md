---
title: Plugins
description: Learn what a Porter plugin can do and see a listing of available plugins
---

The Porter client is extensible and anyone can write a plugin to integrate with
Porter. Plugins extend the Porter client, reimplementing Porter's default
functionality.

For example, Porter saves installation data, credential sets and
parameter sets using the [mongodb-docker plugin], which is
suitable for development and testing. Porter also includes a [mongdb plugin],
which connects to a remote MongoDB server using a configured connection string,
which is intended for production use. You could write your own plugin to better
integrate with a MongoDB as a Server offering from your cloud provider.

[Plugins are very different from mixins][vs], which give you building blocks for
authoring bundles. There are a couple [types of plugins][types] and a single
plugin binary may contain multiple implementations.

# Available Plugins

Below are plugins that are either maintained by the Porter authors, or are community mixins that are known to be well-maintained.
Use the [porter plugins search](/cli/porter_plugins_search) command to see all known plugins.

See the [Search Guide][search-guide] on how to search for available plugins and/or
add your own to the list.

[mongodb plugin]: /plugins/mongodb/
[mongodb-docker plugin]: /plugins/mongodb-docker/

[vs]: /mixins-vs-plugins/
[types]: /plugins/types/
[search-guide]: /package-search/
