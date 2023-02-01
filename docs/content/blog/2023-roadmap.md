---
title: "Porter's 2023 Roadmap"
description: "What's on the docket for Porter in 2023"
date: "2023-01-31"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://carolynvanslyck.com/"
authorimage: "https://github.com/carolynvs.png"
tags: ["roadmap"]
---

Now that 2022 is finally over and [Porter v1 is out the door](https://getporter.org/blog/v1-is-here/), I'd like to share our plans for this coming year.
<!--more-->

You can find a review of 2022 and our high level plans for 2023 in our [Porter Annual Review for 2022](https://github.com/cncf/toc/pull/951/files) but let's go over the highlights for a sneek peak of what's to come.

* [Porter Operator](#porter-operator)
* [Advanced Dependencies](#advanced-dependencies)
* [Improved Bundle Security](#improved-bundle-security)
* [Signing and SBOMs](#signing-and-sboms)

### Porter Operator

The [Porter Operator](/operator/) is a Kubernetes operator that runs bundles on your cluster. With the operator you can:

* Automate running bundles on Kubernetes
* Upgrade to new versions of bundles when they are released
* Integrate bundle deployments into your existing pipelines

Many users, especially adopters with large-scale Porter deployments, are eagerly awaiting managing their installations using a GitOps based workflow with the Porter Operator and Flux. We made great progress on the operator last year and our [v1 milestone](https://github.com/getporter/operator/milestone/1) outlines the remaining work necessary for a stable release.

### Advanced Dependencies

Porter plans to support advanced workflows and dependency scenarios ([PEP003](https://github.com/getporter/proposals/blob/main/pep/003-dependency-namespaces-and-labels.md)), where users can define bundles that have a complex dependency graph.
Dependencies may be an interface such as requiring a MySQL database whether it came from a dedicated server, a Helm chart, or a database as a service from a cloud provider.
Dependencies may also be resolved to existing installations of bundles, such as a shared dev database or a redis instance for the staging environment.

This allows better reuse of bundles, so that a bundle author can write a bundle for just their software, and they don't need to write cloud provider specific bundles, such as "my software on Azure", "my software on Amazon", etc.
Instead, they can create a bundle that requires another bundle that matches a specified interface and Porter can satisfy that dependency differently depending on the target environment.

Development is already in-progress and a lot of initial groundwork in the form of spec changes and refactoring has already been merged into main to support the next version of dependencies.

### Improved Bundle Security

Mixins are wrappers around common deployment tools, such as Terraform and Helm, that make it easier to use that tool inside a bundle with Porter.
Currently, they are distributed as binaries, and embedded inside the bundle's image.

We want to improve bundle security by distributing and executing **mixins as bundles**, essentially decomposing bundles into smaller bundles that we can more securely distribute and execute ([PEP005](https://github.com/getporter/proposals/blob/main/pep/005-mixins-are-bundles.md)).
This has a number of advantages such as:

  * Secure distribution mechanism of mixins. 
    This allows us to no longer need to "install" mixins, or trust those installed mixins.
    Bundles can reference mixins from OCI registries and take advantage of existing security tooling and software to distribute, scan, and sign mixins.
  * Isolate credentials used by a bundle to just the relevant portions of the bundle.
    For example, only the Azure mixin should have access to your Azure credentials, and only the Helm mixin should have access to your kubeconfig.
    Credentials applicable to one mixin are not exposed to other unrelated mixins or scripts in the bundle.
  * Achieve significant performance improvements due to improved reuse of cached image layers.

All these features rely upon the Advanced Dependencies feature, so these will be delivered after that work is completed.

## Signing and SBOMs

There has been huge leaps forward in the world of signing and software bill of materials in the past year that we are ready to jump on and use.
New versions of the OCI Distribution spec, alternatives to Notary such as cosign, and a plethora of SBOM generating libraries are available for us to build into Porter.

Our plan is to incrementally add features allowing you to:

* Sign bundles
* Distribute the bundle signature when the bundle is published
* Validate a bundle signature
* Generate a software bill of materials for possibly multiple aspects of a bundle: the bundle itself, the mixins and tools used in the bundle, and of course the software deployed by the bundle.

We plan to build support for these features incrementally throughout the year, so you won't see everything get rolled out at once in a big bang release.

## Contributors Welcome!
Phew! This is a big chunk of work for a pretty small group of maintainers and contributors.
So if you find any of this interesting, please consider helping us out! 

Our [Contributing](https://getporter.org/contribute/) landing page has details on how to get involved with the project.
We welcome all types of contributions, such as:

* Feedback on use cases and requirements
* Testing out early releases of features
* Reviews of our designs and proposals
* Saying hi in meetings and reminding us that someone cares about any of this shipping
* Happy emoji reactions on our slack announcements and progress reports

This is going to be a big year, and we hope you come along for the ride!
