---
title: Work on Porter with us! 💖
description: We are putting out the welcome mat for new contributors! Learn how to get started as a contributor and work your way up to a maintainer.
---

✅ Are you interested in contributing to an open-source Go project?

🌈 Do you care about being in a welcoming, inclusive community?

🚀 Would you like to get into a project at the beginning and have an impact?

We are seeking out new contributors with the hopes of building up both
casual contributors and enticing some of you into becoming reviewers and
maintainers.

<p align="right">
Carolyn Van Slyck<br/>
✨ Chief Emoji Offer<br/>
Co-creator of Porter<br/>
</p>

# Getting Started

* [What is the project?](#what-is-the-project)
* [What can you work on?](#what-can-you-work-on)
* [Who can be a maintainer?](#who-can-be-a-maintainer)

## What is the project?

Porter is a command-line tool written in Go that implements the [Cloud Native
Application Bundle Specification](https://deislabs.io/cnab). Bundles package up
not just your application, but also the tools and logic needed to deploy and
manage it.

There are a lot of ways to make a bundle. Porter is designed to make robust
bundles with your existing tools and scripts from your pipeline. Already using
helm, terraform and bash? Perfect! Porter [glues][glue] that together for you into a
CNAB bundle. Our goal is a great developer experience.

Porter uses the concept of mixins to build existing tools into bundles. Here is
a [demo][demo] that deploys to Digital Ocean and Kubernetes
using the Terraform and Helm mixins.

[glue]: https://carolynvanslyck.com/blog/2019/04/porter/
[demo]: https://youtu.be/ciA1YuGOIo4

## Where should you start?

Every new contributor should read our [Code of Conduct][conduct] and use our
[Contributing Guide][contributing] to understand what to expect when
contributing to our repositories. The guide also covers the project's code
structure, makefile commands, how to preview documentation and other useful
things to know.

The contributing guide explains how to [find an issue][find-an-issue]. We do use
two labels: [good first issues][good-first-issue] for new contributors and [help
wanted][help-wanted] issues for our other contributors.

* `good first issue` has extra information to help you make your first contribution.
* `help wanted` are issues suitable for someone who isn't a maintainer and usually 
   also has extra guidance.

[conduct]: /src/CODE_OF_CONDUCT.md
[contributing]: /src/CONTRIBUTING.md
[find-an-issue]: /src/CONTRIBUTING.md#find-an-issue
[good-first-issue]: /board/good+first+issue
[help-wanted]: /board/help+wanted

## What can you work on?

We need help with everything! 😊 Whether you are new to Go, or are a
cloud-native gopher expert, we have stuff for you to do.

* Add commands to the porter cli. This is work that never ends and is suitable
  for all levels of gophers.
* Create a mixin! You can start use the [Porter Skeletor][skeletor] repository
  as a template to start, along with the [Mixin Developer Guide][mixin-dev-guide].
* Coordinate between writing an upstream CNAB specification, such as security or 
  dependencies, and implementing it in Porter.
* Implement new runtimes so that Porter can work inside Kubernetes, and other 
  virtualization providers.

The [roadmap][roadmap] will give you a good idea of the larger features that we
are working on right now. That may help you decide what you would like to work
on after you have tackled an issue or two to learn how to contribute to Porter.
If you would like to contribute regularly to a larger issue on the roadmap,
reach out to a maintainer on [Slack][slack].

[skeletor]: https://github.com/deislabs/porter-skeletor
[mixin-dev-guide]: /mixin-dev-guide/
[roadmap]: /roadmap

## Who can be a maintainer?

Porter is not a Microsoft-only project. Anyone can not only contribute but
work their way up the [contribution ladder][ladder] from contributor to 
maintainer to admin.

<p align="center">Sound like fun? 🙋🏽‍♀️ Join us!</p>

[ladder]: /src/CONTRIBUTION_LADDER.md
[slack]: /community#slack
