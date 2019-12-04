# Contributing Guide

---
* [How to help](#how-to-help)
  * [Find an issue](#find-an-issue)
  * [When to open a pull request](#when-to-open-a-pull-request)
  * [How to get your pull request reviewed fast](#how-to-get-your-pull-request-reviewed-fast)
  * [The life of a pull request](#the-life-of-a-pull-request)
* [Contribution Ladder](#contribution-ladder)
  * [Contributor](#contributor)
  * [How to become a contributor](#how-to-become-a-contributor)
  * [Maintainer](#maintainer)
  * [How to become a maintainer](#how-to-become-a-maintainer)
  * [Admin](#admin)
  * [How to become an admin](#admin)
* [Developer Tasks](#developer-tasks)
  * [Initial setup](#initial-setup)
  * [Makefile explained](#makefile-explained)
  * [Install mixins](#install-mixins)
  * [Preview documentation](#preview-documentation)
* [Code structure and practices](#code-structure-and-practices)
  * [What is the general code layout?](#what-is-the-general-code-layout)
  * [Logging](#logging)

---

# How to help

We welcome your contributions and participation! If you aren't sure what to
expect, here are some norms for our project so you feel more comfortable with
how things will go.

## Code of Conduct

The Porter community is governed by our [Code of Conduct](CODE_OF_CONDUCT.md).
This includes but isn't limited to: the porter and related mixin repositories,
slack, interactions on social media, project meetings, conferences and meetups.

## Find an issue

We have [good first issues][good-first-issue] for new contributors and [help
wanted][help-wanted] issues for our other contributors.

* `good first issue` has extra information to help you make your first contribution.
* `help wanted` are issues suitable for someone who isn't a core maintainer.

Maintainers will do our best regularly make new issues for you to solve and then
help out as you work on them. 💖

We have a [roadmap](README.md#roadmap) that will give you a good idea of the
larger features that we are working on right now. That may help you decide what
you would like to work on after you have tackled an issue or two to learn how to
contribute to Porter. If you would like to contribute regularly to a larger
issue on the roadmap, reach out to a maintainer on [Slack][slack].

Another great way to contribute is to create a mixin! You can start use the
[Porter Skeletor][skeletor] repository as a template to start, along with the
[Mixin Developer Guide][mixin-dev-guide].

[skeletor]: https://github.com/deislabs/porter-skeletor
[mixin-dev-guide]: https://porter.sh/mixin-dev-guide/

## When to open a pull request

It's OK to submit a PR directly for problems such as misspellings or other
things where the motivation/problem is unambiguous.

If there isn't an issue for your PR, please make an issue first and explain the
problem or motivation for the change you are proposing. When the solution isn't
straightforward, for example "Implement missing command X", then also outline
your proposed solution. Your PR will go smoother if the solution is agreed upon
before you've spent a lot of time implementing it.

Since Porter is a CLI, the "solution" will usually look like this:

```console
$ porter newcommand [OPTIONAL] [--someflag VALUE]
example output
```

## How to test your pull request

We recommend running the following every time:

```
make verify build test-unit
```

If your test modified anything related to running a bundle, also run:

```
make test-integration
```

If you want to know _all_ the targets that the CI runs, look at
<build/azure-pipelines.pr-automatic.yml> and <build/azure-pipelines.pr-manual.yml>.

## How to get your pull request reviewed fast

🚧 If you aren't done yet, create a draft pull request or put WIP in the title
so that reviewers wait for you to finish before commenting.

1️⃣ Limit your pull request to a single task. Don't tackle multiple unrelated
things, especially refactoring. If you need large refactoring for your change,
chat with a maintainer first, then do it in a separate PR first without any
functionality changes.

🎳 Group related changes into commits will help us out a bunch when reviewing!
For example, when you change dependencies and check in vendor, do that in a
separate commit.

😅 Make requested changes in new commits. Please don't ammend or rebase commits
that we have already reviewed. When your pull request is ready to merge, you can
rebase your commits yourself, or we can squash when we merge. Just let us know
what you are more comfortable with.

🚀 We encourage [follow-on PRs](#follow-on-pr) and a reviewer may let you know in
their comment if it is okay for their suggestion to be done in a follow-on PR.
You can decide to make the change in the current PR immediately, or agree to
tackle it in a reasonable amount of time in a subsequent pull request. If you
can't get to it soon, please create an issue and link to it from the pull
request comment so that we don't collectively forget.

## The life of a pull request

1. You create a draft or WIP pull request. Reviewers will ignore it mostly
   unless you mention someone and ask for help. Feel free to open one and use
   the pull request to see if the CI passes. Once you are ready for a review,
   remove the WIP or click "Ready for Review" and leave a comment that it's
   ready for review.

   If you create a regular pull request, a reviewer won't wait to review it.
1. A reviewer will assign themselves to the pull request. If you don't see
   anyone assigned after 3 business days, you can leave a comment asking for a
   review, or ping in [slack][slack]. Sometimes we have busy days, sick days,
   weekends and vacations, so a little patience is appreciated! 🙇‍♀️
1. The reviewer will leave feedback.
    * `nits`: These are suggestions that you may decide incorporate into your pull
      request or not without further comment.
    * It can help to put a 👍 on comments that you have implemented so that you
      can keep track.
    * It is okay to clarify if you are being told to make a change or if it is a
      suggestion.
1. After you have made the changes (in new commits please!), leave a comment. If
   3 business days go by with no review, it is okay to bump.
1. When a pull request has been approved, the reviewer will squash and merge
   your commits. If you prefer to rebase your own commits, at any time leave a
   comment on the pull request to let them know that.

At this point your changes are available in the [canary][canary] release of Porter!

[canary]: https://porter.sh/install/#canary

### Follow-on PR

A follow-on PR is a pull request that finishes up suggestions from another pull
request.

When the core of your changes are good, and it won't hurt to do more of the
changes later, our preference is to merge early, and keep working on it in a
subsequent. This allows us to start testing out the changes in our canary
builds, and more importantly enables other developers to immediately start
building their work on top of yours.

This helps us avoid pull requests to rely on other pull requests. It also avoids
pull requests that last for months, and in general we try to not let "perfect be
the enemy of the good". It's no fun to watch your work sit in purgatory, and it
kills contributor momentum.

# Contribution Ladder

Our ladder defines the roles and responsibilities on this project and how to
participate with the goal of moving from a user to a maintainer. You will need
to gain people's trust, demonstrate your competence and understanding, and meet
the requirements of the role.

## Community Members

Everyone is a community member! 😄 You've read this far so you are already ahead. 💯

Here are some ideas for how you can be more involved and participate in the community:

* Comment on an issue that you’re interested in.
* Submit a pull request to fix an issue.
* Report a bug.
* Share a bundle that you made and how it went.
* Come chat with us in [Slack][slack].

They must follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Contributor

[Contributors][contributors] have the following capabilities:

* Have issues and pull requests assigned to them
* Apply labels, milestones and projects
* [Mark issues as duplicates](https://help.github.com/en/articles/about-duplicate-issues-and-pull-requests)
* Close, reopen, and assign issues and pull requests

They must agree to and follow this Contributing Guide.

### How to become a contributor

To become a contributor, the maintainers of the project would like to see you:

* Comment on issues with your experiences and opinions.
* Add your comments and reviews on pull requests.
* Contribute pull requests.
* Open issues with bugs, experience reports, and questions.

Contributors and maintainers will do their best to watch for community members
who may make good contributors. But don’t be shy, if you feel that this is you,
please reach out to one or more of the contributors or maintainers.

[contributors]: https://github.com/orgs/deislabs/teams/porter-contributors

## Maintainer

[Maintainers][maintainers] are members with extra capabilities:

* Be a [Code Owner](.github/CODEOWNERS) and have reviews automatically requested.
* Review pull requests.
* Merge pull requests.

Maintainers also have additional responsibilities beyond just merging code:

* Help foster a safe and welcoming environment for all project participants.
  This will include understanding and enforcing our [Code of Conduct](CODE_OF_CONDUCT.md).
* Organize and promote pull request reviews, e.g. prompting community members,
  contributors, and other maintainers to review.
* Triage issues, e.g. adding labels, promoting discussions, finalizing decisions.
* Help organize our development meetings, e.g. schedule, organize and
  execute agenda.

They must agree to and follow the [Reviewing Guide](REVIEWING.md).

[maintainers]: https://github.com/orgs/deislabs/teams/porter-maintainers

### How to become a maintainer

To become a maintainer, we would like you to see you be an effective
contributor, and show that you can do some of the things maintainers do.
Maintainers will do their best to regularly discuss promoting contributors. But
don’t be shy, if you feel that this is you, please reach out to one or more of
the maintainers.

## Admin

[Admins][admins] are maintainers with extra responsibilities:

* Create new mixin repositories
* Manage porter-* repositories
* Manage porter-* teams

[admins]: https://github.com/orgs/deislabs/teams/porter-admins

### How to become an admin

It isn't expected that all maintainers will need or want to move up to admin. If
you are a maintainer, and find yourself often asking an admin to do certain
tasks for you and you would like to help out with administrative tasks, please
reach out to one or more of the admins.

# Developer Tasks

## Initial setup

1. Clone this repository with `go get -u get.porter.sh/porter`. Porter relies on being in the GOPATH.
1. Run `make build install`.

You now have canary builds of porter and all the mixins installed.

## Makefile explained

Here are the most common Makefile tasks

* `build` builds all binaries, porter and internal mixins.
* `build-porter-client` just builds the porter client for your operating
  system. It does not build the porter-runtime binary. Useful when you just want
  to do a build and don't remember the proper way to call `go build` yourself.
* `build-porter` builds both the porter client and runtime. It does not clean
   up generated files created by packr, so you usually want to also run `clean-packr`.
* `install-porter` installs just porter from your bin into **/usr/local/bin**.
* `install-mixins` installs just the mixins from your bin into
  **/usr/local/bin**. This is useful when you are working on the exec or
  kubernetes mixin.
* `install` installs porter _and_ the mixins from your bin into **/usr/local/bin**.
* `test-unit` runs the unit tests.
* `test-integration` runs the integration tests. This requires a kubernetes
  cluster setup with credentials located at **~/.kube/config**. Expect this to
  take 10 minutes.
* `test-cli` runs a small test of end-to-end tests that require a kubernetes
  cluster (same as `test-integration`).
* `docs-preview` hosts the docs site. See [Preview
  Documentation](#preview-documentation).
* `test` runs all the tests.
* `clean-packr` removes extra packr files that were a side-effect of the build.
  Normally this is run automatically but if you run into issues with packr and
  dep, run this command.
* `dep-ensure` runs dep ensure for you while taking care of packr properly. Use
  this if your PRs are often failing on `verify-vendor` because of packr. This
  can be avoided entirely if you use `make build-porter-client` or `make build`.
* `verify-vendor` cleans up packr generated files and verifies that dep's Gopkg.lock
   and vendor/ are up-to-date. Use this makefile target instead of running
   dep check manually.

## Install mixins

When you run `make build`, the canary\* build of mixins are automatically
installed into your bin directory in the root of the repository. You can use
`porter mixin install NAME` to install the latest released version of a mixin.

\* canary = most recent successful build of master

## Preview documentation

We use [Hugo](gohugo.io) to build our documentation site, and it is hosted on
[Netlify](netlify.com).

1. [Install Hugo](https://gohugo.io/getting-started/installing) using `brew install hugo`,
`choco install hugo` or `go get -u github.com/gohugoio/hugo`.
1. Run `make docs-preview` to start Hugo. It will watch the file system for changes.
1. Open <http://localhost:1313> to preview the site.

If anyone is interested in contributing changes to our makefile to improve the
authoring experience, such as doing this with Docker so that you don't need Hugo
installed, it would be a welcome contribution! ❤️

[good-first-issue]: https://github.com/deislabs/porter/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22
[help-wanted]: https://github.com/deislabs/porter/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22

## Command Documentation

Our commands are documented at <https://porter.sh/cli> and that documentation is
generated by our CLI. You should regenerate that documentation when you change
any files in **cmd/porter** by running `make docs-gen` which is run every time
you run `make build`.

# Code structure and practices

Carolyn Van Slyck gave a talk about the design of Porter, [Designing
Command-Line Tools People Love][porter-design] that you may find helpful in
understanding the why's behind its command grammar, package structure, use of
dependency injection and testing strategies.

[porter-design]: https://carolynvanslyck.com/talks/#gocli

## What is the general code layout?

* **cmd**: go here to add a new command or flag to porter or one of the mixins in
  this repository
* **docs**: our website
* **pkg**
  * **build**: implements building the invocation image.
  * **cache**: handles the cache of bundles that have been pulled by commands
  like `porter install --tag`.
  * **cnab**: deals with the CNAB spec
    * **cnab-to-oci**: talking to an OCI registry.
    * **config-adapter**: converting porter.yaml to bundle.json.
    * **extensions**: extensions to the CNAB spec, at this point that's just
  dependencies.
    * **provider**: the CNAB runtime, i.e. `porter install`.
  * **config**: anything related to `porter.yaml` and `~/.porter`.
  * **context**: essentially dependency injection that's needed throughout Porter,
    such as stdout, stderr, stdin, filesystem and command execution.
  * **exec**: the exec mixin
  * **kubernetes**: the kubernetes mixin
  * **mixin**: enums, functions and interfaces for the mixin framework.
    * **feed**: works with mixin atom feeds
    * **provider**: handles communicating with mixins
  * **porter**: the implementation of the porter commands. Every command in Porter
    has a corresponding function in here.
      packr
    * **version**: reusable library used by all the mixins for implementing their
  * **templates**: files that need to be compiled into the porter binary with
      version command.
* **scripts**:
  * **install**: Porter [installation](https://porter.sh/install) scripts
* **tests** have Go-based integration tests.
* **vendor** we use dep and check in vendor.

## Logging

**Print to the `Out` property for informational messages and send debug messages to the `Err` property.**

Example:

```golang
fmt.Fprintln(p.Out, "Initiating battlestar protocol")
fmt.Fprintln(p.Err, "DEBUG: loading plans from r2d2...")
```

Most of the structs in Porter have an embedded
`github.com/deislabs/porter/pkg/context.Context` struct. This has both `Out` and
`Err` which represent stdout and stderr respectively. You should log to those
instead of directly to stdout/stderr because that is how we capture output in
our unit tests. That means use `fmt.Fprint*` instead of `fmt.Print*` so that you
can pass in `Out` or `Err`.

Some of our commands are designed to be consumed by another tool and intermixing
debug lines and the command output would make the resulting output unusable. For
example, `porter schema` outputs a json schema and if log lines were sent to
stdout as well, then the resulting json schema would be unparsable. This is why
we send regular command output to `Out` and debug information to `Err`. It
allows us to then run the command and see the debug output separately, like so
`porter schema --debug 2> err.log`.

[slack]: https://porter.sh/community#slack
