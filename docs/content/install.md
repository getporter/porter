---
title: Install Porter
description: Installing the Porter client and mixins
---

> Join our [mailing list] for announcements of releases and announcements of new features.
> Connect with other Porter users and contributors on [Slack].

We have a few release types available for you to use:

* [Latest](#latest)
* [Canary](#canary)
* [Older Version](#older-version)

You can also install and manage [mixins](#mixins) and [plugins](#plugins) using
porter, and use the [Porter VS Code Extension][vscode-ext] to help author
bundles.

All the scripts for Porter v0.37.3+ support [customizing the installation through parameters](#install-script-parameters).

[vscode-ext]: https://marketplace.visualstudio.com/items?itemName=ms-kubernetes-tools.porter-vscode
[ps-link]: https://www.howtogeek.com/126469/how-to-create-a-powershell-profile/
[mailing list]: https://groups.io/g/porter
[Slack]: /community/#slack

# Latest

Install the most recent stable release of porter and its default [mixins](#mixins).

## Latest MacOS
```
curl -L https://cdn.porter.sh/latest/install-mac.sh | bash
```

## Latest Linux
```
curl -L https://cdn.porter.sh/latest/install-linux.sh | bash
```

## Latest Windows
You will need to create a [PowerShell Profile][ps-link] if you do not have one.

```
iwr "https://cdn.porter.sh/latest/install-windows.ps1" -UseBasicParsing | iex
```

# Canary

Install the most recent build from the "main" branch of porter and its [mixins](#mixins).

This saves you the trouble of cloning and building porter and its mixin
repositories yourself. The build may not be stable but it will have new features
that we are developing.

## Canary MacOS
```
curl -L https://cdn.porter.sh/canary/install-mac.sh | bash
```

## Canary Linux
```
curl -L https://cdn.porter.sh/canary/install-linux.sh | bash
```

## Canary Windows
You will need to create a [PowerShell Profile][ps-link] if you do not have one.

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
curl -L https://cdn.porter.sh/$VERSION/install-mac.sh | bash
```

## Older Version Linux
```
VERSION="v0.18.1-beta.2"
curl -L https://cdn.porter.sh/$VERSION/install-linux.sh | bash
```

## Older Version Windows
You will need to create a [PowerShell Profile][ps-link] if you do not have one.

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


[releases]: https://github.com/getporter/porter/releases



# Install Script Parameters

The installation scripts provide the following parameters. Parameters can be specified with environment variables for the macOS and Linux scripts, and on Windows they are named parameters in the script.

## PORTER_HOME

Location where Porter is installed (defaults to ~/.porter).

**Posix Shells**
```bash
export PORTER_HOME=/alt/porter/home
curl -L REPLACE_WITH_INSTALL_URL | bash
```

**PowerShell**
```powershell
iwr REPLACE_WITH_INSTALL_URL -OutFile install-porter.ps1 -UseBasicParsing
.\install-porter.ps1 -PORTER_HOME C:\alt\porter\home
```

## PORTER_MIRROR

Base URL where Porter assets, such as binaries and atom feeds, are downloaded.
This lets you set up an internal mirror. Note that atom feeds and index files
should be updated in the mirror to point to the mirrored location. Porter does
not alter the contents of these files.

**Posix Shells**
```bash
export PORTER_MIRROR=https://example.com/porter
curl -L REPLACE_WITH_INSTALL_URL | bash
```

**PowerShell**
```powershell
iwr REPLACE_WITH_INSTALL_URL -OutFile install-porter.ps1 -UseBasicParsing
.\install-porter.ps1 -PORTER_MIRROR https://example.com/porter
```

### URL Structure

Configuring a mirror of Porter's assets is out of scope of this document.
Reach out on the Porter [mailing list] for assistance.

Below is the general structure for Porter's asset URLs:

```
PERMALINK/
  - install-linux.sh
  - install-mac.sh
  - install-windows.ps1
  - porter-GOOS-GOARCH[FILE_EXT]
mixins/
  - atom.xml
  - index.json
  - MIXIN/PERMALINK/MIXIN-GOOS-GOARCH[FILE_EXT]
plugins/
  - atom.xml
  - index.json
  - PLUGIN/PERMALINK/PLUGIN-GOOS-GOARCH[FILE_EXT]
```
