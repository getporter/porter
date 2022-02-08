---
title: Install Porter
description: Installing the Porter client and mixins
---

> Join our [mailing list] for announcements of releases and announcements of new features.
> Connect with other Porter users and contributors on [Slack].

We have a few release types available for you to use:

* [Latest](#latest)
* [Canary](#canary)
* [Prerelease](#prerelease)
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

Install the most recent stable release of porter and the [exec mixin].

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

Install the most recent build from the "main" branch of porter and the [exec mixin].

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

# Prerelease

We would love for you to try out [v1 prerelease] and send us any feedback that you have!
Keep in mind that prereleases are not suitable for production workloads. Data migrations will not be provided or supported for prereleases.
Prereleases are intended for you to try out potential new features in Porter and provide feedback about the direction of the feature. They won't work with existing installations.

You can try out different versions of Porter without impacting your current version of Porter by installing to a different location via a modified PORTER_HOME environment variable.

**MacOS**

```bash
export PORTER_HOME=~/.porterv1
export VERSION="v1.0.0-alpha.5"
curl -L https://cdn.porter.sh/$VERSION/install-mac.sh | bash
```

After installing the prerelease, you can switch your current shell session to use the prerelease by setting the PORTER_HOME environment variable and prepending that location to your PATH environment variable.

```bash
export PORTER_HOME=~/.porterv1
export PATH=$PORTER_HOME:$PATH
# Check that you are using the desired version of porter
porter version
```

**Linux**

```bash
export PORTER_HOME=~/.porterv1
export VERSION="v1.0.0-alpha.5"
curl -L https://cdn.porter.sh/$VERSION/install-linux.sh | bash
```

After installing the prerelease, you can switch your current shell session to use the prerelease by setting the PORTER_HOME environment variable and prepending that location to your PATH environment variable.

```bash
export PORTER_HOME=~/.porterv1
export PATH=$PORTER_HOME:$PATH
# Check that you are using the desired version of porter
porter version
```

**Windows**

```powershell
$PORTER_HOME="$env:USERPROFILE\.porterv1"
$VERSION="v1.0.0-alpha.5"
(New-Object System.Net.WebClient).DownloadFile("https://cdn.porter.sh/$VERSION/install-windows.ps1", "install-porter.ps1")
.\install-porter.ps1 -PORTER_HOME $PORTER_HOME
```

After installing the prerelease, you can switch your current shell session to use the prerelease by setting the PORTER_HOME environment variable and prepending that location to your PATH environment variable.

```powershell
$env:PORTER_HOME="$env:USERPROFILE\.porterv1"
$env:PATH+=";$env:PORTER_HOME"
# Check that you are using the desired version of porter
porter version
```

# Older Version

Install an older version of porter, starting with `v0.18.1-beta.2`.
Porter v1.0.0+ only installs porter and the [exec mixin].
Older versions of Porter installed more mixins by default. 

If you need a specific version of a mixin, use the `--version` flag when [installing the mixin](#mixins).

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

We have a number of [mixins](/mixins) to help you get started.
Only the [exec mixin] is installed with Porter, other mixins should be installed separately.

You can update an existing mixin, or install a new mixin using the `porter mixin
install` command:

```console
$ porter mixin install terraform
installed terraform mixin v0.3.0-beta.1 (0d24b85)
```

All of the Porter-authored mixins are published to `https://cdn.porter.sh/mixins/atom.xml`.

# Plugins

We have a couple [plugins](/plugins) which extend Porter and integrate with other cloud providers and software.

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

# Command Completion

Porter provides autocompletion support for Bash, Fish, Zsh, and PowerShell.

> If you use Bash the completion script depends on Bash v4.1 or newer and bash-completion v2.

> The default version for macOS is Bash v3.2 and bash-completion v1. The completion command will not work properly with these versions.
> The Kubernetes project has detailed information for upgrading Bash and installing bash-completion [here].

[here]: https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/#enable-shell-autocompletion

### Initial Setup
The initial setup is to generate a completion script file and have your shell environment source it when you start your shell.

 The completion command will generate its output to standard out and you can capture the output into a file. This file should be put in a place where your shell reads completion files.

An example for Bash:
```bash
porter completion bash > /usr/local/etc/bash_completion.d/porter
```

Once your completion script file is in place you will have to source it for your current shell or start a new shell session.


### Completion Usage

To list available commands for Porter, in your terminal run
```console
$ porter [tab][tab]
```

To find a specific command that starts with _bu_
```console
$ porter bu[tab][tab]

build    bundles
```
Commands that have sub-commands will be displayed with completions as well

```console
$ porter credentials [tab][tab]

delete    edit    generate    list    show
```

> Note: Completion commands are available for Porter's built in commands and flags, future plans include dynamic completion for your project.

[exec mixin]: /mixins/exec/
[v1 prerelease]: /tags/v1/
