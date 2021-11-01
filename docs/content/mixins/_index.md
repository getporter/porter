---
title: Mixins
description: Quickly integrate with existing tools using Porter mixins
aliases:
- /using-mixins/
- /use-mixins/
---

Mixins make Porter special. They are the building blocks that you use when authoring bundles. Find them, use them, [create your own](/mixin-dev-guide/).

Mixins are adapters between the Porter and an existing tool or system. They know how to talk to Porter to include everything
they need to run, such as a CLI or config files, and how to execute their steps in the Porter manifest.

There are [many mixins](/mixins/) created both by the Porter Authors and the community.
Only the [exec mixin](/mixins/exec/) is installed by default.


# Next

* [Use a mixin in a bundle](/author-bundles/#mixins)
* [Mixin Architecture](/mixin-dev-guide/architecture/)

# Available Mixins

Below are mixins that are either maintained by the Porter authors, or are community mixins that are known to be well-maintained.
Use the [porter mixins search](/cli/porter_mixins_search) command to see all known mixins.

See the [Search Guide][search-guide] on how to search for available plugins and/or
add your own to the list.

[search-guide]: /package-search/
