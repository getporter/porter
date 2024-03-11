---
title: "Quickstart"
description: ""
weight: 2
---

This guide covers the most commonly used actions of Porter (install, upgrade, uninstall) and navigates users through their first Bundle experience. ✨ 

## Pre-requisites
- (Docker)[https://docs.docker.com/get-docker/]

### Install Porter

#### Latest MacOS

```
curl -L https://cdn.porter.sh/latest/install-mac.sh | bash
```

#### Latest Linux

```
curl -L https://cdn.porter.sh/latest/install-linux.sh | bash
```

#### Latest Windows

You will need to create a [PowerShell Profile][ps-link] if you do not have one.

```
iwr "https://cdn.porter.sh/latest/install-windows.ps1" -UseBasicParsing | iex
```

## Install a Bundle

To install a bundle, you use the `porter install` command.

```
porter install porter-hello --reference ghcr.io/getporter/examples/porter-hello:v0.2.0

executing install action from examples/porter-hello (installation: /)
Install Hello World
Hello, porter
execution completed successfully!
```

In this example, you are installing the v0.2.0 version of the ghcr.io/getporter/examples/porter-hello bundle from its location in the default registry (Docker Hub) and setting the installation name to porter-hello.

## List Bundle Installations

To see the list of bundle installations, use the `porter list` command.

```console
$ porter list
NAME              CREATED          MODIFIED         LAST ACTION   LAST STATUS
porter-hello      21 minutes ago   21 minutes ago   install       succeeded
```

In this example, it shows the bundle metadata along with the creation time, modification time, the last action that was performed, and the status of the last action.


## Show Installation Information

To see information about an installation, use the `porter show` command with the name of the installation.

```console
$ porter show porter-hello
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

```
porter upgrade

executing upgrade action from examples/porter-hello (installation: /porter-hello)
World 2.0
Hello, porter
execution completed successfully!
```


## Cleanup

To clean up the resources installed from the bundle, use the `porter uninstall` command.

```
porter uninstall porter-hello
```

## Next Steps

You've learned how to install, upgrade, and uninstall a bundle.
From here, you can dive into what a bundle is - or create your own!

- [Next: What is a Bundle?](/quickstart/bundles/)
- [Next: Create my own bundle](/getting-started/create-a-bundle/)
- [Learn more about use cases for bundles](/learning/#the-devil-is-in-the-deployments-bundle-use-cases)


