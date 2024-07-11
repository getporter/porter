---
title: "Quickstart"
description: ""
weight: 2
aliases:
  - /quickstart/
---

This guide covers the most commonly used actions of Porter (install, upgrade, uninstall) and navigates users through their first Bundle experience. ✨ 

## Pre-requisites
- [Docker](https://docs.docker.com/get-docker/)

- Porter

### Install Porter

##### MacOS

```
curl -L https://cdn.porter.sh/latest/install-mac.sh | bash
```

##### Linux

```
curl -L https://cdn.porter.sh/latest/install-linux.sh | bash
```

##### Windows

```
iwr "https://cdn.porter.sh/latest/install-windows.ps1" -UseBasicParsing | iex
```
You will need to create a PowerShell Profile if you do not have one.


## Install a Bundle

To install a bundle, you use the `porter install` command.

```console
$ porter install porter-hello --reference ghcr.io/getporter/examples/porter-hello:v0.2.0
```

```
executing install action from examples/porter-hello (installation: /)
Install Hello World
Hello, porter
execution completed successfully!
```

This example installs version `0.2.0` of the `ghcr.io/getporter/examples/porter-hello` bundle from the GitHub container registry. The installation name is `porter-hello`.


## List Bundle Installations

To see the list of bundle installations, use the `porter list` command.

```console
$ porter list
```

```
NAME              CREATED          MODIFIED         LAST ACTION   LAST STATUS
porter-hello      21 minutes ago   21 minutes ago   install       succeeded
```

This `porter list` example shows bundle metadata such as the bundle installation name, creation and modification times, the last action and its status.


## Show Installation Information

To see information about an installation, use the `porter show` command with the name of the installation.

```console
$ porter show porter-hello
```
```
Name: hello
Bundle: ghcr.io/getporter/examples/porter-hello
Version: 0.2.0
Created: 2021-05-24
Modified: 2021-05-24

History:
------------------------------------------------------------------------
  Run ID                      Action   Timestamp   Status     Has Logs
------------------------------------------------------------------------
  01F1SVDSQDVKGC0VAABZE9ERQK  install  2021-03-27  failed     true
  01F1SVVRGSWG3FKY2ZATN4XTKC  install  2021-03-27  succeeded  true
```

## Upgrade the Installation

To upgrade the resources managed by the bundle, use `porter upgrade`.

```console
$ porter upgrade porter-hello
```

```
executing upgrade action from examples/porter-hello (installation: /porter-hello)
World 2.0
Hello, porter
execution completed successfully!
```


## Cleanup

To clean up the resources installed from the bundle, use the `porter uninstall` command.

```console
$ porter uninstall porter-hello
```

## Next Steps

Now you've seen the basics to install, upgrade or uninstall a bundle.
From here, you can dive into what a bundle is - or create your own!

- [Next: What is a Bundle?](/docs/quickstart/bundles/)
- [Next: Create my own bundle](/docs/getting-started/create-bundle/)
- [Learn more about use cases for bundles](/docs/learn/#the-devil-is-in-the-deployments-bundle-use-cases)


