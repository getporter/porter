---
title: Plugins
description: Learn what a Porter plugin can do and see a listing of available plugins
---

The Porter client is extensible and anyone can write a plugin to integrate with
Porter. Plugins extend the Porter client, reimplementing Porter's default
functionality. For example, Porter saves installation data, credential sets and
parameter sets using the local filesystem to ~/.porter by default. A plugin can
change that behavior to save them to cloud storage instead.

[Plugins are very different from mixins][vs], which give you building blocks for
authoring bundles. There are a couple [types of plugins][types] and a single
plugin binary may contain multiple implementations.

See the [Search Guide][search-guide] on how to search for available plugins and/or
add your own to the list.

[vs]: (/mixins-vs-plugins/)
[types]: (/plugins/types/)
[search-guide]: (/package-search/)
