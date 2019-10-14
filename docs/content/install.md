---
title: Install Porter
description: Installing the Porter client and mixins
---

We have a few release types available for you to use:

* [Latest](#latest)
* [Canary](#canary)
* [Older Version](#older-version)

You can also install and manage [mixins](#mixins) using porter.

# Latest

Install the most recent stable release of porter and its default [mixins](#mixins).

## Latest MacOS
```
curl https://cdn.deislabs.io/porter/latest/install-mac.sh | bash
```

## Latest Linux
```
curl https://cdn.deislabs.io/porter/latest/install-linux.sh | bash
```

## Latest Windows
```
iwr "https://cdn.deislabs.io/porter/latest/install-windows.ps1" -UseBasicParsing | iex
```

# Canary

Install the most recent build from master of porter and its [mixins](#mixins).

This saves you the trouble of cloning and building porter and its mixin
repositories yourself. The build may not be stable but it will have new features
that we are developing.

## Canary MacOS
```
curl https://cdn.deislabs.io/porter/canary/install-mac.sh | bash
```

## Canary Linux
```
curl https://cdn.deislabs.io/porter/canary/install-linux.sh | bash
```

## Canary Windows
```
iwr "https://cdn.deislabs.io/porter/canary/install-windows.ps1" -UseBasicParsing | iex
```

# Older Version

Install an older version of porter. This also installs the latest version of all the mixins.
If you need a specific version of a mixin, use the `--version` flag when
[installing the mixin](#mixin).

See the porter [releases][releases] page for a list of older porter versions.

## Older Version MacOS
```
curl https://cdn.deislabs.io/porter/VERSION/install-mac.sh | bash
```

## Older Version Linux
```
curl https://cdn.deislabs.io/porter/VERSION/install-linux.sh | bash
```

## Older Version Windows
```
iwr "https://cdn.deislabs.io/porter/VERSION/install-windows.ps1" -UseBasicParsing | iex
```

# Mixins

We have a number of [mixins](/mixins) to help you get started. The stable ones
are installed by default:

* exec
* kubernetes
* helm
* azure
* terraform

You can update an existing mixin, or install a new mixin using the `porter mixin
install` command:

```console
$ porter mixin install terraform
installed terraform mixin
v0.3.0-beta.1 (0d24b85)
```

All of the DeisLabs mixins are published to `https://cdn.deislabs.io/porter/atom.xml`.

[releases]: https://github.com/deislabs/porter/releases
