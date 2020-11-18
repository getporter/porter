---
title: Work on Porter with us! 💖
description: We are putting out the welcome mat for new contributors! Learn how to get started as a contributor and work your way up to a maintainer.
layout: single
---

✅ Are you interested in contributing to an open-source Go project?

🌈 Do you care about being in a welcoming, inclusive community?

🚀 Would you like to get into a project at the beginning and have an impact?

We are seeking out new contributors with the hopes of building up both
casual contributors and enticing some of you into becoming reviewers and
maintainers.

# Getting Started

{{< toc >}}

## What is the project?


[Porter] is a command-line tool written in Go that implements the [Cloud
Native Application Bundle Specification](https://deislabs.io/cnab). Package your
application artifact, client tools, configuration and deployment logic together
as a versioned bundle that you can distribute, and then install with a single
command.

<p align=center>Try our <a href="/quickstart/">QuickStart</a> to learn more about Porter!</p>

[Porter]: /

## Where should you start?

Every new contributor should read our [Code of Conduct][conduct] and use our
[Contributing Guide][contributing] to understand what to expect when
contributing to our repositories. The guide also covers how to get the source
code, build and test it, and then submit your first pull request.

We have a [tutorial] to get your environment setup, make your first change
to Porter, and try it out.

The contributing guide explains how to [find an issue]. We do use
two labels:

* [good first issue] has extra information to help you make your first contribution.
* [help wanted] are issues suitable for someone who isn't a core maintainer.

[conduct]: /src/CODE_OF_CONDUCT.md
[contributing]: /contribute/guide/
[find an issue]: /contribute/guide/#find-an-issue
[good first issue]: /board/good+first+issue
[help wanted]: /board/help+wanted
[tutorial]: /contribute/tutorial/

## What can you work on?

We need help with everything! 😊 Whether you are new to Go or cloud-native
gopher expert, are interested in design or writing, we have stuff for you to
do:

* Add commands to the porter cli. This is work that never ends and is suitable
  for all levels of gophers.
* Create a mixin! You can start use the [Porter Skeletor][skeletor] repository
  as a template to start, along with the [Mixin Developer Guide][mixin-dev-guide].
  Here's the list of [existing mixins] and [requested mixins].
* Coordinate between writing an upstream CNAB specification, such as security or 
  dependencies, and implementing it in Porter.
* Implement new runtimes so that Porter can work inside Kubernetes, and other 
  virtualization providers.
* Improve our website's design by contributing diagrams, graphics, improved layouts,
  organizing the information.
* Fill in gaps in our documentation by copying answers from Slack and GitHub into 
  our FAQ, or creating new pages and content.
* Project management or other skillsets would be amazing as well! Contact
  us on the [mailing list] or [Slack]. and let's coordinate. 🙌

🙋🏻‍♀️ If this is your first open-source project, we have opportunities for you to
learn in a safe space! So if you aren't sure where to start, or need a mentor
to help guide you through learning the project and how to contribute, say 
"I'm new to OSS and would like a mentor" in our [Slack] channel and a maintainer 
will reach out to you and get you set up.

The [roadmap] will give you a good idea of the larger features that we
are working on right now. That may help you decide what you would like to work
on after you have tackled an issue or two to learn how to contribute to Porter.
If you would like to contribute regularly to a larger issue on the roadmap,
contact us on the [mailing list] or [Slack].

[skeletor]: https://github.com/getporter/skeletor
[mixin-dev-guide]: /mixin-dev-guide/
[roadmap]: /roadmap
[existing mixins]: https://github.com/getporter/packages/blob/main/mixins/index.json
[requested mixins]: https://github.com/getporter/porter/issues?q=is%3Aissue+is%3Aopen+label%3A%22mixin+idea%22
[mailing list]: https://groups.io/g/porter 

## Who can be a maintainer?

Porter is not a Microsoft-only project. Anyone can not only contribute but
work their way up the [contribution ladder] from contributor to 
maintainer to admin.

<p align="center">Sound like fun? 👍 Join us!</p>

[contribution ladder]: /src/CONTRIBUTION_LADDER.md
[Slack]: /community#slack
