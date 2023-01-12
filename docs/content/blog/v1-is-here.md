---
title: "Porter v1.0.0 is here!"
description: "Announcing Porter's v1.0.0 release, what is new, and our plans going forward"
date: "2022-10-03"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://carolynvanslyck.com/"
authorimage: "https://github.com/carolynvs.png"
tags: ["roadmap", "v1", "release-notes"]
---

Porter v1.0.0 is finally here! 🎉
<!--more-->

## What is Porter?

[Porter](/) takes everything you need to do a deployment—command-line tools, configuration files, secrets, bash scripts, and the application itself—and glues it all together.
Then Porter packages that into a versioned bundle distributed over standard OCI registries or plain tgz files.
With Porter, anyone can install your application without needing deep knowledge of your deployment process or underlying tech stack.

Learn more about Porter:

* [QuickStart: Use a bundle](/quickstart)
* [Learn when bundles make sense, when they don’t, and what your day could look like if you were using them](/learning/#the-devil-is-in-the-deployments-bundle-use-cases)
* [Get a high level overview of bundles and Porter](/architecture/)
* [Understand Porter's security features](/security-features/)

## What's new in v1?

If you haven't been following Porter's v1 pre-releases, below are some of the new features introduced since v0.38.

### Buildkit Support

Porter now uses Docker Buildkit instead of legacy Docker to build images.
This dramatically speeds up the build process and allows you to take advantage of features such as build arguments, improved layer caching, new Dockerfile syntax, mounting secrets and ssh connections into the image during build and more.

### Installations

Porter now understands the concept of installations.
You can see the current definition of an installation, the bundle it's using, its version, the associated parameters and credentials, logs from previous runs, and metadata about the installation's current state.
Installations can be isolated in namespaces and have labels applied to them to make it easier to manage multiple installations and environments.

Porter also supports specifying the desired state of an installation (or credential set, parameter set).
Porter will reconcile that against its database and determine the appropriate action to execute, if any, to reach that state.
This makes it much easier to automate running Porter based on triggers such as a git push, or when a new version of a bundle is released, without having to deal with figuring out if you should call install or upgrade, and other things that Porter can figure out for you.

Desired state also supports the upcoming [Porter Operator](/operator/), a Kubernetes operator that automates running Porter on a Kubernetes cluster, similar to the Helm Operator.
The operator is still under development but is coming up on our roadmap now that Porter v1 is live.

### Improved Data Persistence

Previously, we supported Azure Blob Storage, Azure Tables, and storing data as files in your home directory.
Porter now stores its data in Mongodb, for improved query performance when working with larger databases.
You can use Mongo Atlas, Azure CosmosDB with Mongo API, or any Mongodb instance with Porter.

In addition to changing Porter's database, we have eliminated sensitive data from the database altogether.
Porter has always resolved and injected secrets into a bundle Just-In-Time from a secret store such as Hashicorp Vault or Azure Key Vault.
With Porter v1, when we need to work with sensitive data, it is [persisted back to your configured secret store](/blog/persist-sensitive-data-safely/).
This includes sensitive data in your config file, such as tokens or connection strings.
Your configuration file can now [load sensitive data directly from your configured secret store](/blog/secret-free-config/) so that your database or config file can't be used gain access to sensitive data.

### Tighter Security

* We have lots of bug fixes around connecting to OCI registries and remote Docker daemons both in production environments and in development.
* Porter resolves referenced image tags to digests when the bundle is built, ensuring that your bundle pins its dependencies.
* Bundles no longer run as root.
* Porter images are now redistributed on the [PlatformOne IronBank registry](https://p1.dso.mil/products/iron-bank).
  These images are built on an isolated network, regularly scanned for CVEs, and new releases are available on average 1-2 days after the official Porter releases on GitHub.
* Porter publish checks for an existing bundle at the destination location to avoid accidentally overwriting a published bundle.
  You can still force push a bundle, but it's opt-in now.

Read more about Porter's [security features](/security-features/).

## What's next?

Our v1 release is not a stopping point, but instead a way point where we can now say "Porter is stable and safe to use in production".
We have big plans going forward, adding new features on top of v1 incrementally:

* **Advanced Dependencies**: Our initial implementation of dependencies was always limited in scope. More complete and powerful dependency support is already underway. Learn more in [PEP003 Advanced Dependencies](https://github.com/getporter/proposals/blob/main/pep/003-dependency-namespaces-and-labels.md).
* **Distribute Mixins as Bundles**: After we have advanced dependency support, we are improving how mixins are distributed and executed so that they are BUNDLES! This will significantly improve performance, layer caching, mixin distribution, and bundle execution security. Learn more in [PEEP005 Mixins are Bundles](https://github.com/getporter/proposals/blob/main/pep/005-mixins-are-bundles.md).
* **Porter Operator v1**: The [Porter Operator] is far enough along for you to try, and we aim to quickly get it ready for a v1 release.
* **Support for signing bundles**: Porter will support integration with Notary for signing and verifying bundles.

[Porter Operator]: /operator/
