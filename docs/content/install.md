---
title: Install Porter
description: Installing the Porter client and mixins
---

> Join our [mailing list] for announcements of releases and announcements of new features.
> Connect with other Porter users and contributors on [Slack].

* [Install or Upgrade Porter](#install-or-upgrade)
* [Clean Install](#clean-install)
* [Install Latest Release](#latest)
* [Install a Canary Build](#canary)
* [Install a Mixin](#mixins)
* [Install a Plugin](#plugins)
* [Install the Porter VS Code Extension][vscode-ext]
* [Customize the Installation Script](#install-script-parameters)
* [Configure Command Completion](#command-completion)

# Install or Upgrade

The examples below use a hard-coded version of Porter. There may be a newer version available which you can check for on our [release] page.
Set VERSION to the most recent [release] version number.

**MacOS**

```bash
export VERSION="v1.0.1"
curl -L https://cdn.porter.sh/$VERSION/install-mac.sh | bash
```

**Linux**

```bash
export VERSION="v1.0.1"
curl -L https://cdn.porter.sh/$VERSION/install-linux.sh | bash
```

**Windows**

```powershell
$VERSION="v1.0.1"
(New-Object System.Net.WebClient).DownloadFile("https://cdn.porter.sh/$VERSION/install-windows.ps1", "install-porter.ps1")
.\install-porter.ps1
```

## Running multiple versions 

If you have multiple versions of Porter installed on the same machine, you can switch between then by setting the PORTER_HOME environment variable and adding the desired version of Porter to your PATH.

**Bash**

```bash
export PORTER_HOME=~/.porterv1
export PATH=$PORTER_HOME:$PATH
# Check that you are using the desired version of porter
porter version
```

**Windows**

```powershell
$env:PORTER_HOME="$env:USERPROFILE\.porterv1"
$env:PATH+=";$env:PORTER_HOME"
# Check that you are using the desired version of porter
porter version
```

[vscode-ext]: https://marketplace.visualstudio.com/items?itemName=ms-kubernetes-tools.porter-vscode
[ps-link]: https://www.howtogeek.com/126469/how-to-create-a-powershell-profile/
[mailing list]: https://groups.io/g/porter
[Slack]: /community/#slack

# Clean Install

To perform a clean installation of Porter:

1. Remove the PORTER_HOME directory, which by default is located at `~/.porter`.
2. Start over with a fresh database. If you were using an external database, update your porter configuration file to use a different database.

   Otherwise, if you had not specified a storage plugin or database in the configuration file, then your database is located in a container in Docker.
   Remove the mongodb container and volume, so that when Porter is run again, it creates a new database: 
   ```
   docker rm -f porter-mongodb-docker-plugin
   docker volume rm porter-mongodb-docker-plugin-data
   ```
3. Install Porter following the instructions on this page.

# Latest

Install the most recent tagged v1 release of porter and the [exec mixin].

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

Install the most recent build of Porter and the [exec mixin] from the main branch.

This saves you the trouble of cloning and building porter and its mixin
repositories yourself. The build may not be stable, but it will have new features
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

# Mixins

We have a number of [mixins](/mixins) to help you get started.
Only the [exec mixin] is installed with Porter v1.0.0+, other mixins should be installed separately.

You can update an existing mixin, or install a new mixin using the `porter mixin
install` command:

```console
$ porter mixin install terraform --version v1.0.0-rc.1
installed terraform mixin v1.0.0-rc.1 (090064b)
```

All the Porter-authored mixins are published to `https://cdn.porter.sh/mixins/atom.xml`.

# Plugins

We have a couple [plugins](/plugins) which extend Porter and integrate with other cloud providers and software.

You can update an existing plugin, or install a new plugin using the `porter plugin
install` command:

```console
$ porter plugin install azure --version v1.0.0-rc.1
installed azure plugin v1.0.0-rc.1 (d0d98f4)
```

All the Porter-authored plugins are published to `https://cdn.porter.sh/plugins/atom.xml`.


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

> Note: Completion commands are available for Porter's built-in commands and flags, future plans include dynamic completion for your project.

[exec mixin]: /mixins/exec/
[release]: https://github.com/getporter/porter/releases
