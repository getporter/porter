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

[Porter](/) takes everything you need to do a deployment, the application itself and the entire process to deploy it: command-line tools, configuration files, secrets, and bash scripts to glue it all together.
Then packages that into a versioned bundle distributed over standard Docker registries or plain tgz files.
With Porter anyone can install your application without deep knowledge of your deployment process, regardless of the underlying tech stack.

Learn more about Porter:

* [QuickStart: Use a bundle](/quickstart)
* [Learn when bundles make sense, when they don’t, and what your day could look like if you were using them](/learning/#the-devil-is-in-the-deployments-bundle-use-cases)
* [Get a high level overview of bundles and Porter](/architecture/)
* [Understand Porter's security features](/security-features/)

## What's new in v1?

If you haven't been following Porter's v1 pre-releases, here are some of the new features introduced since v0.38:

### Desired State

You can now define the desired state of an installation: what bundle to install, custom parameters to specify and the credentials to give the bundle.
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

Read more about Porter's [security features](/security-features/).

## What's next?

