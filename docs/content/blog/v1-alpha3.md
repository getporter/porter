---
title: "Porter 1.0.0-alpha.3 is here!"
description: "Check out the new features introduced in Porter v1.0.0-alpha.3"
date: "2021-09-21"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://carolynvanslyck.com/"
authorimage: "https://github.com/carolynvs.png"
tags: ["release-notes", "v1"]
---

Check out the new features introduced in Porter v1.0.0-alpha.3 such as MongoDB, namespaces, bundle state and credential import.
<!--more-->

This is a big release for our v1 alpha with LOTS of major changes that we hope make Porter easier to use such as [MongoDB](#mongodb-support), [namespaces](#namespaces), [labels](#labels), [bundle state](#bundle-state), and [credential import](#import-credentials-and-parameter-sets). Our [release notes] have a full accounting of all **fifty-eight** pull requests that were included in this release.

A v1 version of our documentation is available at https://release-v1.porter.sh.

[release notes]: https://github.com/getporter/porter/releases/tag/v1.0.0-alpha.3

## Breaking Changes
Safety first! In this release we have removed support for deprecated flags such as --tag and fields like tag in porter.yaml.
If you have been ignoring a warning message in Porter's output for months, now is the time to fix those messages before upgrading.

## MongoDb Support
Porter now stores its data in MongoDB! We no longer store installation records on the filesystem, though we still do have a ~/.porter directory with our configuration and cache.
In dev and test, Porter by default runs MongoDB in a container for you on porter 27018 and connects to it.
The mongodb-docker plugin stores the underlying mongo data in a Docker volume, but persistence is not guaranteed which
is why this is only for non-production environments.
Once Porter v1 is ready for production, the mongodbF` storage plugin is what you should use to connect to an existing MongoDB server.

```toml
default-storage = "mydb"

[[storage]]
  name = "mydb"
  plugin = "mongodb"

  [storage.config]
    url = "mongodb://username:password@host:port"
```

CosmosDB works as well, though we have a known performance problem with our CosmosDB indices that will be addressed in the next release.

Any existing v0.38 compatible storage mixins, such as the azure storage plugin, do not work with this release.
We are still finalizing the new storage plugin protocol, so we suggest holding off on making a custom storage plugin until its ready.

## Namespaces
Porter resources such as installations, credential sets and parameter sets can optionally be defined within a namespace. 
Resources that are not defined in a namespace are considered global.
When an installation is defined in a namespace, it can reference a credential or parameter set that is also defined in that namespace or at the global scope.
Resources defined globally cannot reference other resources that are defined in a namespace.

You can set the current namespace in the [Porter configuration file](https://release-v1.porter.sh/configuration/#config-file) using the namespace setting.

When an installation references a parameter or credential set, Porter first looks for a resource with that name in the current namespace.
If one does not exist, Porter then looks for that resource at the global level.
This lets you define a common set of credentials to be used for an environment, like staging environment credentials that everyone can reuse, while allowing for overriding that resource within a particular namespace.

```bash
# list just in the current namespace defined in your porter config
porter installations list

# list in the specified namespace
porter installation list --namespace myuser

# list across all namespaces
porter installation list --all-namespaces

# list across all namespaces (alternate syntax)
porter installation list --namespace '*'
```

## Labels
Porter resources also now support labels which may be used for filtering with the list commands.
For example:

```
porter installations list --label env=dev --label app=myapp
porter credentials list --label env=dev
porter parameters list --label env=dev
```

## Bundle State
Bundle authors can now declare that certain files are "state" and should be preserved between bundle runs.
A prime example is the tfstate and tfvars files created by the terraform mixin.
Now instead of using the old method of declaring those files as both parameters and outputs, you should declare them as **state** and Porter will ensure that they are always available to your bundle.

This corrects the problem where if an installation that used the terraform mixin, without a remote backend configured, failed to install, on subsequent install runs the resources originally created by terraform were not recorded and reused, leaving it up to the user to manually cleanup resources.
With state, when install fails and is repeated, the terraform mixin is able to pick up where it left off.

## Import credentials and parameter sets
Use the apply command to import a credential or parameter set from a file.

```
porter credentials apply mycreds.yaml
porter credentials apply myparams.json
```

You can export a credential or parameter set to a file with the show command.

```
porter credentials show mycreds --output yaml > mycreds.yaml
porter credentials show myparams --output json > myparams.json
```

# Install the v1 Prerelease

We would love for you to try out v1.0.0-alpha.3 and send us any feedback that you have! Keep in mind that the v1 prerelease is not suitable for running with production workloads, and that data migrations will not be provided or supported for v1 prerelease.
The prerelease is intended for you to try out the new features in Porter, and provide feedback but won't work with existing installations.

One way to try out Porter without messing with your current installation of Porter is to install Porter into a different 
PORTER_HOME directory.

**MacOS**

```bash
export PORTER_HOME=~/.porterv1
export VERSION="v1.0.0-alpha.3"
curl -L https://cdn.porter.sh/$VERSION/install-mac.sh | bash
```

**Linux**

```bash
PORTER_HOME=~/.porterv1
VERSION="v1.0.0-alpha.3"
curl -L https://cdn.porter.sh/$VERSION/install-linux.sh | bash
```

**Windows**

```powershell
$PORTER_HOME="$env:USERPROFILE\.porterv1"
$VERSION="v1.0.0-alpha.3"
(New-Object System.Net.WebClient).DownloadFile("https://cdn.porter.sh/$VERSION/install-windows.ps1", "install-porter.ps1")
.\install-porter.ps1 -PORTER_HOME $PORTER_HOME
```

Now when you want to use the v1 version of Porter, set the PORTER_HOME environment variable and add it to your PATH.

**Posix Shells**
```bash
export PORTER_HOME=~/.porterv1
export PATH="$PORTER_HOME:$PATH"
```

**PowerShell**
```powershell
$env:PORTER_HOME="$env:USERPROFILE\.porterv1"
$env:PATH+=";$env:PORTER_HOME"
```
