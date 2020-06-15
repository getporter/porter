---
title: Install Porter
description: Installing the Porter client and mixins
---

We have a few release types available for you to use:

* [Latest](#latest)
* [Canary](#canary)
* [Older Version](#older-version)

You can also install and manage [mixins](#mixins) using porter,
and use the [Porter VS Code Extension][vscode-ext] for help
authoring bundles. For Windows installation, you will need to
create a [PowerShell Profile][ps-link] if you do not have one. 

[vscode-ext]: https://marketplace.visualstudio.com/items?itemName=ms-kubernetes-tools.porter-vscode
[ps-link]: https://www.howtogeek.com/126469/how-to-create-a-powershell-profile/

# Latest

Install the most recent stable release of porter and its default [mixins](#mixins).

## Latest MacOS
```
curl https://cdn.porter.sh/latest/install-mac.sh | bash
```

## Latest Linux
```
curl https://cdn.porter.sh/latest/install-linux.sh | bash
```

## Latest Windows
```
iwr "https://cdn.porter.sh/latest/install-windows.ps1" -UseBasicParsing | iex
```

# Canary

Install the most recent build from master of porter and its [mixins](#mixins).

This saves you the trouble of cloning and building porter and its mixin
repositories yourself. The build may not be stable but it will have new features
that we are developing.

## Canary MacOS
```
curl https://cdn.porter.sh/canary/install-mac.sh | bash
```

## Canary Linux
```
curl https://cdn.porter.sh/canary/install-linux.sh | bash
```

## Canary Windows
```
iwr "https://cdn.porter.sh/canary/install-windows.ps1" -UseBasicParsing | iex
```

# Older Version

Install an older version of porter, starting with `v0.18.1-beta.2`. This also
installs the latest version of all the mixins. If you need a specific version of
a mixin, use the `--version` flag when [installing the mixin](#mixins).

See the porter [releases][releases] page for a list of older porter versions.
Set `VERSION` to the version of Porter that you want to install.

## Older Version MacOS
```
VERSION="v0.18.1-beta.2"
curl https://cdn.porter.sh/$VERSION/install-mac.sh | bash
```

## Older Version Linux
```
VERSION="v0.18.1-beta.2"
curl https://cdn.porter.sh/$VERSION/install-linux.sh | bash
```

## Older Version Windows
```
$VERSION="v0.18.1-beta.2"
iwr "https://cdn.porter.sh/$VERSION/install-windows.ps1" -UseBasicParsing | iex
```

# Mixins

We have a number of [mixins](/mixins) to help you get started, and stable mixins
are installed by default.

You can update an existing mixin, or install a new mixin using the `porter mixin
install` command:

```console
$ porter mixin install terraform
installed terraform mixin v0.3.0-beta.1 (0d24b85)
```

All of the Porter-authored mixins are published to `https://cdn.porter.sh/mixins/atom.xml`.

# Plugins

We are working on building out [plugins](/plugins) to extend Porter and the stable
plugins are installed by default.

You can update an existing plugin, or install a new plugin using the `porter plugin
install` command:

```console
$ porter plugin install azure --version canary
installed azure plugin v0.1.1-10-g7071451 (7071451)
```

All of the Porter-authored plugins are published to `https://cdn.porter.sh/plugins/atom.xml`.


[releases]: https://github.com/deislabs/porter/releases

