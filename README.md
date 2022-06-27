<img align="right" src="docs/static/images/porter-docs-header.svg" width="300px" />

[![Build Status](https://dev.azure.com/getporter/porter/_apis/build/status/porter-canary?branchName=main)](https://dev.azure.com/getporter/porter/_build/latest?definitionId=26&branchName=main)

# Porter

Package your application, client tools, configuration, and deployment logic into an installer that you can distribute and run with a single command.
Based on the Cloud Native Application Bundle Specification, [CNAB](https://deislabs.io/cnab), Porter provides a declarative authoring experience that lets you focus on what you know best: your application.

<p align="center">Learn all about Porter at <a href="https://porter.sh">porter.sh</a></p>

# <a name="mixins"></a>Porter Mixins

Mixins provide out-of-the-box support for interacting with different tools and services from inside a bundle. You can always create a mixin, or use the exec mixin and a Custom Dockerfile if a custom mixin doesn't exist yet.

[Porter Mixins](https://porter.sh/mixins/) are available for below platform's:

| Platform                                                                                                                                                                                                                        | Supported?  |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :---------: |
| <img src="docs\static\images\mixins\docker_icon.png" width="20" height="20" vertical-align="middle" /> [Docker](https://porter.sh/mixins/docker/)                                            |     ✔️      |
| <img src="docs\static\images\mixins\docker-compose.png" width="20" height="20" vertical-align="middle" /> [Docker-Compose](https://porter.sh/mixins/docker-compose/)            |     ✔️      |
| <img src="docs\static\images\mixins\kubernetes.svg" width="20" height="20" vertical-align="middle" /> [Kubernetes](https://porter.sh/mixins/kubernetes/)            |     ✔️      |
| <img src="docs\static\images\mixins\helm.svg" width="20" height="20" vertical-align="middle" /> [Helm](https://porter.sh/mixins/helm/)            |     ✔️      |
| <img src="docs\static\images\mixins\gcp.png" width="20" height="20" vertical-align="middle" /> [GCloud](https://porter.sh/mixins/gcloud/)            |     ✔️      |
| <img src="docs\static\images\mixins\terraform_icon.png" width="20" height="20" vertical-align="middle" /> [Terraform](https://porter.sh/mixins/terraform/)            |     ✔️      |
| <img src="docs\static\images\mixins\aws.svg" width="20" height="20" vertical-align="middle" /> [aws](https://porter.sh/mixins/aws/)            |     ✔️      |
| <img src="docs\static\images\plugins\azure.png" width="20" height="20" vertical-align="middle" /> [Azure](https://porter.sh/mixins/azure/)            |     ✔️      |
| <img src="docs\static\images\mixins\exec.png" width="20" height="20" vertical-align="middle" /> [exec](https://porter.sh/mixins/exec/)            |     ✔️      |

# <a name="Plugins"></a>Porter Plugins

Plugins let you store Porter's data and retrieve secrets from an external service.

[Porter Plugins](https://porter.sh/plugins/) are available for below platform's:

| Platform                                                                                                                                                                                                                        | Supported?  |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :---------: |
| <img src="docs\static\images\plugins\hashicorp.png" width="20" height="20" vertical-align="middle" /> [Hashicorp](https://porter.sh/plugins/hashicorp/)                                            |     ✔️      |
| <img src="docs\static\images\plugins\azure.png" width="20" height="20" vertical-align="middle" /> [Azure](https://porter.sh/plugins/azure/)            |     ✔️      |
| <img src="docs\static\images\mixins\kubernetes.svg" width="20" height="20" vertical-align="middle" /> [Kubernetes](https://porter.sh/plugins/kubernetes/)            |     ✔️      |


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

[Mailing List]: https://porter.sh/mailing-list
[Slack]: https://porter.sh/community/#slack
[Open an Issue]: https://github.com/getporter/porter/issues/new/choose
[Forum]: https://porter.sh/forum/
[Dev Meeting]: https://porter.sh/community/#dev-meeting
[Porter Enhancement Proposals]: https://porter.sh/contribute/proposals/

# Looking for Contributors

Want to work on Porter with us? 💖 We are actively seeking out new contributors
with the hopes of building up both casual contributors and enticing some of you
into becoming reviewers and maintainers.

<p align="center">Start with our <a href="https://porter.sh/contribute/">New Contributors Guide</a>

Porter wouldn't be possible without our [contributors][contributors], carrying
the load and making it better every day! 🙇‍♀️

[contributors]: /CONTRIBUTORS.md

# Do you use Porter?

Take our [user survey](https://porter.sh/user-survey) and let us know if you are using Porter.
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

<p align="center">Check out our <a href="https://porter.sh/roadmap">roadmap</a></p>

Our [version strategy] explains how we version the project, when you should expect
breaking changes in a release, and the process for the v1 release.

[board]: https://porter.sh/board
[version strategy]: https://porter.sh/project/version-strategy/
