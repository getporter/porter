---
title: Architecture
description: Detailed look at how Porter does its magic 🎩✨
---

Porter is an implementation of the [Cloud Native Application Bundle](/cnab/) specification and creates installers, known as bundles, that understand how to install not only your application but its infrastructure and configuration.
A bundle helps you package up the logic for installing your application so that you can hand it off to another team, a customer, or just a co-worker who doesn't know all the ins and outs of your application and provide them a consistent installation experience that doesn't require them to know what tools you use to deploy or how the sausage is made.

## Buildtime

A bundle author is responsible for understanding how to deploy an application and automate it inside a bundle using existing deployment tools, such as terraform or helm.
They use [mixins] to install tools into the bundle, and can include additional files like helm charts, kustomize files, or terraform modules.
In the porter.yaml file, they automate each step of the application's deployment:

* Collecting credentials
* Using parameters to customize the installation
* Creating infrastructure
* Setting up configuration files
* Installing software

The bundle author then builds the bundle into an OCI artifact which includes a docker image with the necessary files and development tools along with the bundle's metadata and any other docker images used by the bundle.
Once a bundle is built, it can be distributed using Docker / OCI registries.
This allows you to use existing tools and infrastructure to share your application with other teams, customers, and end-users.

Because the bundle contains everything you need to deploy, included referenced images, you can even [move a bundle into a disconnected or airgapped environment](/administrators/airgap/).
When written with airgap deployments in mind, a bundle can be deployed anywhere without requiring access to the original network or the internet.

Learn more about [how Porter works at buildtime](/architecture-buildtime/).

## Runtime

End users can then discover, inspect and run bundles with a consistent interface (porter) that doesn't rely on them understanding the underlying set of tools, scripts and architecture of the application.
Regardless of what is being installed, the commands look the same:

```
porter install --credential-set USER_CREDS --parameter-set CUSTOM_PARAMETERS --reference BUNDLE_REFERENCE
```

Once a bundle is installed, is tracked as an "installation" of the bundle.
An installation is a record of which bundle was installed, the parameters used to customize the installation, previous runs of the bundle, the current version of the bundle, along with other useful metadata.

Learn more about [how Porter works at runtime](/architecture-runtime/).

## See Also

* [Security Features](/security-features/)
* [Create a Bundle](/bundle/create/)
* [Distribute Bundles](/distribute-bundles/)
* [Airgapped Deployments](/administrators/airgap/)

[mixins]: /mixins/
