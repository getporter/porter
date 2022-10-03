<img align="right" src="docs/static/images/porter-docs-header.svg" width="300px" />

[![Build Status](https://dev.azure.com/getporter/porter/_apis/build/status/porter-canary?branchName=main)](https://dev.azure.com/getporter/porter/_build/latest?definitionId=26&branchName=main)
<a href="https://getporter.org/find-issue" alt="Find an issue to work on">
<img src="https://img.shields.io/github/issues-search?label=%22help%20wanted%22%20issues&query=org%3Agetporter%20label%3A%22good%20first%20issue%22%2C%22help%20wanted%22%20no%3Aassignee" /></a>

# Porter

Package your application, client tools, configuration, and deployment logic into an installer that you can distribute and run with a single command.
Based on the Cloud Native Application Bundle Specification, [CNAB](https://deislabs.io/cnab), Porter provides a declarative authoring experience that lets you focus on what you know best: your application.

<p align="center">Learn all about Porter at <a href="https://getporter.org">getporter.org</a></p>

# <a name="mixins"></a>Porter Mixins

Mixins provide out-of-the-box support for interacting with different tools and services from inside a bundle. You can always create a mixin, or use the exec mixin and a Custom Dockerfile if a custom mixin doesn't exist yet.

[Porter Mixins](https://getporter.org/mixins/) are available for below platform's:

| Platform                                                                                                                                                                                                                        | Supported?  |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :---------: |
| <img src="docs\static\images\mixins\docker_icon.png" width="20" height="20" vertical-align="middle" /> [Docker](https://getporter.org/mixins/docker/)                                            |     ✔️      |
| <img src="docs\static\images\mixins\docker-compose.png" width="20" height="20" vertical-align="middle" /> [Docker-Compose](https://getporter.org/mixins/docker-compose/)            |     ✔️      |
| <img src="docs\static\images\mixins\kubernetes.svg" width="20" height="20" vertical-align="middle" /> [Kubernetes](https://getporter.org/mixins/kubernetes/)            |     ✔️      |
| <img src="docs\static\images\mixins\helm.svg" width="20" height="20" vertical-align="middle" /> [Helm](https://getporter.org/mixins/helm/)            |     ✔️      |
| <img src="docs\static\images\mixins\gcp.png" width="20" height="20" vertical-align="middle" /> [GCloud](https://getporter.org/mixins/gcloud/)            |     ✔️      |
| <img src="docs\static\images\mixins\terraform_icon.png" width="20" height="20" vertical-align="middle" /> [Terraform](https://getporter.org/mixins/terraform/)            |     ✔️      |
| <img src="docs\static\images\mixins\aws.svg" width="20" height="20" vertical-align="middle" /> [aws](https://getporter.org/mixins/aws/)            |     ✔️      |
| <img src="docs\static\images\plugins\azure.png" width="20" height="20" vertical-align="middle" /> [Azure](https://getporter.org/mixins/azure/)            |     ✔️      |
| <img src="docs\static\images\mixins\exec.png" width="20" height="20" vertical-align="middle" /> [exec](https://getporter.org/mixins/exec/)            |     ✔️      |

# <a name="Plugins"></a>Porter Plugins

Plugins let you store Porter's data and retrieve secrets from an external service.

[Porter Plugins](https://getporter.org/plugins/) are available for below platform's:

| Platform                                                                                                                                                                                                                        | Supported?  |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :---------: |
| <img src="docs\static\images\plugins\hashicorp.png" width="20" height="20" vertical-align="middle" /> [Hashicorp](https://getporter.org/plugins/hashicorp/)                                            |     ✔️      |
| <img src="docs\static\images\plugins\azure.png" width="20" height="20" vertical-align="middle" /> [Azure](https://getporter.org/plugins/azure/)            |     ✔️      |
| <img src="docs\static\images\mixins\kubernetes.svg" width="20" height="20" vertical-align="middle" /> [Kubernetes](https://getporter.org/plugins/kubernetes/)            |     ✔️      |


# Contact

* [Mailing List] - Great for following the project at a high level because it is
  low traffic, mostly release notes and blog posts on new features.
* [Forum] - Share an idea or propose a design where everyone can benefit from
  the discussion and find answers to questions later.
* [Dev Meeting] - Biweekly meeting where we discuss [Porter Enhancement Proposals], demo new features and help other contributors.
* [Open an Issue] - If you are having trouble or found a bug, ask on GitHub so
  that we can prioritize it and make sure you get an answer.
* [Slack] - We have a #porter channel and there's also #cnab for deep thoughts
  about the CNAB specification.

[Mailing List]: https://getporter.org/mailing-list
[Slack]: https://getporter.org/community/#slack
[Open an Issue]: https://github.com/getporter/porter/issues/new/choose
[Forum]: https://getporter.org/forum/
[Dev Meeting]: https://getporter.org/community/#dev-meeting
[Porter Enhancement Proposals]: https://getporter.org/contribute/proposals/

# Looking for Contributors

Want to work on Porter with us? 💖 We are actively seeking out new contributors
with the hopes of building up both casual contributors and enticing some of you
into becoming reviewers and maintainers.

<p align="center">Start with our <a href="https://getporter.org/contribute/">New Contributors Guide</a>

Porter wouldn't be possible without our [contributors][contributors], carrying
the load and making it better every day! 🙇‍♀️

[contributors]: /CONTRIBUTORS.md

# Do you use Porter?

Take our [user survey](https://getporter.org/user-survey) and let us know if you are using Porter.
Project funding is contingent upon knowing that we have active users!

# Roadmap

Porter is an open-source project and things get done as quickly as we have
motivated contributors working on features that interest them. 😉

We use a single [project board][board] across all of our repositories to track
open issues and pull requests.

The roadmap represents what the maintainers have said that they are
currently working on and plan to work on over the next few months. We use the
"on-hold" bucket to communicate items of interest that do not have a
maintainer who will be working on them.

<p align="center">Check out our <a href="https://getporter.org/roadmap">roadmap</a></p>

Our [version strategy] explains how we version the project, when you should expect
breaking changes in a release, and the process for the v1 release.

[board]: https://getporter.org/board
[version strategy]: https://getporter.org/project/version-strategy/
