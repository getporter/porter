# Contributing Guide

---
* [How to help](#how-to-help)
  * [Find an issue](#find-an-issue)
  * [Which branch to use](#which-branch-to-use)  
  * [When to open a pull request](#when-to-open-a-pull-request)
  * [How to get your pull request reviewed fast](#how-to-get-your-pull-request-reviewed-fast)
  * [Signing your commits](#signing-your-commits)
  * [The life of a pull request](#the-life-of-a-pull-request)
* [Contribution Ladder](#contribution-ladder)
* [Developer Tasks](#developer-tasks)
  * [Initial setup](#initial-setup)
  * [Makefile explained](#makefile-explained)
  * [Install mixins](#install-mixins)
  * [Preview documentation](#preview-documentation)
  * [View a trace of a Porter command](#view-a-trace-of-a-porter-command)
  * [Write a blog post](#write-a-blog-post)
* [Code structure and practices](#code-structure-and-practices)
  * [What is the general code layout?](#what-is-the-general-code-layout)
  * [Logging](#logging)
  * [Breaking Changes](#breaking-changes)
* [Infrastructure](#infrastructure)
  * [CDN Setup](#cdn-setup)
  * [Custom Windows CI Agent](#custom-windows-ci-agent)
  * [Releases](#releases)

---

# How to help

We welcome your contributions and participation! If you aren't sure what to
expect, here are some norms for our project so you feel more comfortable with
how things will go.

If this is your first contribution to Porter, we have a [tutorial] that walks you
through how to setup your developer environment, make a change and test it.

[tutorial]: https://getporter.org/contribute/tutorial/

## Code of Conduct

The Porter community is governed by our [Code of Conduct][coc].
This includes but isn't limited to: the porter and related mixin repositories,
slack, interactions on social media, project meetings, conferences and meetups.

[coc]: https://getporter.org/src/CODE_OF_CONDUCT.md

## Find an issue

Use the [getporter.org/find-issue] link to find good first issues for new contributors and help wanted issues for our other contributors.

When you have been contributing for a while, take a look at the "Backlog" column on our [project board][board] for high priority issues.
The project board is at the organization level, so it contains issues from across all the Porter repositories. 

* [`good first issues`][good-first-issue] has extra information to help you make your first contribution.
* [`help wanted`][help-wanted] are issues suitable for someone who isn't a core maintainer.
* `hmm 🛑🤔` issues should be avoided. They are not ready to be worked on yet
  because they are not finished being designed or we aren't sure if we want the
  feature, etc.

Maintainers will do their best to regularly make new issues for you to solve and then 
help out as you work on them. 💖

We have a [roadmap] that will give you a good idea of the
larger features that we are working on right now. That may help you decide what
you would like to work on after you have tackled an issue or two to learn how to
contribute to Porter. If you have a big idea for Porter, learn [how to propose
a change to Porter][pep].

Another great way to contribute is to create a mixin! You can start using the
[Porter Skeletor][skeletor] repository as a template to start, along with the
[Mixin Developer Guide][mixin-dev-guide].

When you create your first pull request, add your name to the bottom of our 
[Contributors][contributors] list. Thank you for making Porter better! 🙇‍♀️

[getporter.org/find-issue]: https://getporter.org/find-issue/
[contributors]: https://getporter.org/src/CONTRIBUTORS.md                                          
[skeletor]: https://github.com/getporter/skeletor
[mixin-dev-guide]: https://getporter.org/mixin-dev-guide/
[good-first-issue]: https://getporter.org/board/good+first+issue
[help-wanted]: https://getporter.org/board/help+wanted
[board]: https://getporter.org/board
[slack]: https://getporter.org/community#slack
[roadmap]: https://getporter.org/src/README.md#roadmap
[pep]: https://getporter.org/contribute/proposals/

## Which branch to use

Unless the issue specifically mentions a branch, please created your feature branch from the release/v1 branch.

For example:

```
# Make sure you have the most recent changes to release/v1
git checkout release/v1
git pull

# Create a branch based on release/v1 named MY_FEATURE_BRANCH
git checkout -b MY_FEATURE_BRANCH
```

## When to open a pull request

It's OK to submit a PR directly for problems such as misspellings or other
things where the motivation/problem is unambiguous.

If there isn't an issue for your PR, please make an issue first and explain the
problem or motivation for the change you are proposing. When the solution isn't
straightforward, for example, "Implement missing command X", then also outline
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
mage Build TestUnit
```

If your test modified anything related to running a bundle, also run:

```
mage TestIntegration
```

If you want to know _all_ the targets that the CI runs, look at
<build/azure-pipelines.pr-automatic.yml>.

## How to get your pull request reviewed fast

🚧 If you aren't done yet, create a draft pull request or put WIP in the title
so that reviewers wait for you to finish before commenting.

1️⃣ Limit your pull request to a single task. Don't tackle multiple unrelated
things, especially refactoring. If you need large refactoring for your change,
chat with a maintainer first, then do it in a separate PR first without any
functionality changes.

🎳 Group related changes into separate commits to make it easier to review. 

😅 Make requested changes in new commits. Please don't amend or rebase commits
that we have already reviewed. When your pull request is ready to merge, you can
rebase your commits yourself, or we can squash when we merge. Just let us know
what you are more comfortable with.

🚀 We encourage [follow-on PRs](#follow-on-pr) and a reviewer may let you know in
their comment if it is okay for their suggestion to be done in a follow-on PR.
You can decide to make the change in the current PR immediately, or agree to
tackle it in a reasonable amount of time in a subsequent pull request. If you
can't get to it soon, please create an issue and link to it from the pull
request comment so that we don't collectively forget.

## Signing your commits

You can automatically sign your commits to meet the DCO requirement for this
project by running the following command: `make setup-dco`.

Licensing is important to open source projects. It provides some assurances that
the software will continue to be available based under the terms that the
author(s) desired. We require that contributors sign off on commits submitted to
our project's repositories. The [Developer Certificate of Origin
(DCO)](https://developercertificate.org/) is a way to certify that you wrote and
have the right to contribute the code you are submitting to the project.

You sign-off by adding the following to your commit messages:

```
Author: Your Name <your.name@example.com>
Date:   Thu Feb 2 11:41:15 2018 -0800

    This is my commit message

    Signed-off-by: Your Name <your.name@example.com>
```

Notice the `Author` and `Signed-off-by` lines match. If they don't, the PR will
be rejected by the automated DCO check.

Git has a `-s` command line option to do this automatically:

    git commit -s -m 'This is my commit message'

If you forgot to do this and have not yet pushed your changes to the remote
repository, you can amend your commit with the sign-off by running 

    git commit --amend -s

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
    * `nits`: These are suggestions that you may decide to incorporate into your pull
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

At this point your changes are available in the [canary][canary] release of
Porter! After your first pull request is merged, you will be invited to the
[Contributors team] which you may choose to accept (or not). Joining the team lets
you have issues in GitHub assigned to you.

[canary]: https://getporter.org/install/#canary
[Contributors team]: https://github.com/orgs/getporter/teams/contributors

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

Our [contribution ladder][ladder] defines the roles and responsibilities for this
project and how to participate with the goal of moving from a user to a
maintainer.

[ladder]: https://getporter.org/src/CONTRIBUTION_LADDER.md

# Developer Tasks

## Initial setup

We have a [tutorial] that walks you through how to set up your developer
environment, make a change and test it.

Here are the key steps, if you run into trouble, the tutorial has more details:

1. Install Go version 1.17 or higher.
1. Clone this repository with `git clone https://github.com/getporter/porter.git ~/go/src/get.porter.sh/porter`.
1. Run `mage Build Install` from within the newly cloned repository.

If you are planning on contributing back to the project, you'll need to
[fork](https://guides.github.com/activities/forking/) and clone your fork. If
you want to build porter from scratch, you can follow the process above and
clone directly from the project.

You now have canary builds of porter and all the mixins installed.

## Makefile explained

🚧 We are in the process of transitioning from make to [mage](https://magefile.org).

### Mage Targets

Below are the targets that have been migrated to mage. Our new contributor
tutorial explains how to [install mage](/contribute/tutorial/#install-mage).

Mage targets are not case-sensitive, but in our docs we use camel case to make
it easier to read. You can run either `mage TestSmoke` or `mage testsmoke` for
example.

* **Build** builds all binaries, porter and internal mixins.
  * **BuildClient** just builds the porter client for your operating system.
    It does not build the porter-runtime binary. Useful when you just want to do a
    build and don't remember the proper way to call `go build` yourself.
  * **BuildPorter**     builds both the porter client and runtime.
* **Clean** removes artifacts from previous builds and test runs.
* **UpdateTestfiles** updates the "golden" test files to match the latest test output.
  This is mostly useful for when you change the schema of porter.yaml which will
  break TestPorter_PrintManifestSchema. Run this target to fix it.
  Learn more about [golden files].
* **Test** runs all the tests.
  * **TestUnit** runs the unit tests
  * **TestSmoke** runs a small suite of tests using the Porter CLI to validate
    that Porter is (mostly) working.
  * **TestIntegration** runs our integration tests, which run the bundles
    against a test KIND cluster.
* **Install** installs porter _and_ the mixins from source into **$(HOME)/.porter/**.
* **DocsPreview** hosts the docs site. See [Preview Documentation](#preview-documentation).
* **DocsGen** generates the CLI documentation for the website. This is run automatically by build.

[golden files]: https://ieftimov.com/post/testing-in-go-golden-files/

### Make Targets

Below are the most common developer tasks. Run a target with `make TARGET`, e.g.
`make setup-dco`.

* `setup-dco` installs a git commit hook that automatically signsoff your commit
  messages per the DCO requirement.

## Test Porter

We have a few different kinds of tests in Porter. You can run all tests types
with `mage test`.

### Unit Tests
 
```
mage TestUnit
```

Should not rely on Docker, or try to really run bundles without key components
mocked. Most structs have test functions, e.g. `porter.NewTestPorter` that are
appropriate for unit tests.

Fast! 🏎💨 This takes about 15s - 3 minutes, depending on your computer hardware.

### Integration Tests

```
mage TestIntegration
```

These tests run parts of Porter, using the Porter structs instead of the cli.
They can use Docker, expect that a cluster is available, etc. These tests all
use functions like `porter.SetupIntegrationTest()` to update the underlying
components so that they hit the real filesystem, and don't mock out stuff like
Docker.

You must have Docker on your computer to run these tests. The test setup handles
creating a Kubernetes cluster and Docker registry. Since they are slow, it is
perfectly fine to not run these locally and rely on the CI build that's triggered
when you push commits to your pull request instead.

When I am troubleshooting an integration test, I will run just the single test
locally by using `go test -run TESTNAME ./...`. If the test needs infrastructure, 
we have scripts that you can use, like `mage StartDockerRegistry` or 
`mage EnsureTestCluster`.

Slow! 🐢 This takes between 8-16 minutes, depending on your computer hardware.

### Smoke Tests

```
mage testSmoke
```

Smoke tests test Porter using the CLI and quickly identify big problems with a
build that would make it unusable.

Short! We want this to always be something you can run in under 3 minutes.

## Install mixins

When you run `mage build`, the canary\* build of mixins are automatically
installed into your bin directory in the root of the repository. You can use
`porter mixin install NAME` to install the latest released version of a mixin.

\* canary = most recent successful build of the "main" branch

## Plugin Debugging

If you are developing a [plugin](https://getporter.org/plugins/) and you want to
debug it follow these steps:

The plugin to be debugged should be compiled and placed in porters plugin path
(e.g. in the Azure plugin case the plugin would be copied to
$PORTER_HOME/plugins/azure/.

The following environment variables should be set:

`PORTER_RUN_PLUGIN_IN_DEBUGGER` should be set to the name of the plugin to be
debugged (e.g. secrets.azure.keyvault to debug the azure secrets plugin)  
`PORTER_DEBUGGER_PORT` should be set to the port number where the delve API will
listen, if not set it defaults to 2345  
`PORTER_PLUGIN_WORKING_DIRECTORY` should be the path to the directory containing
*source code* for the plugin being executed.  

When porter is run it will start delve and attach it to the plugin process, this
exposes the delve API so that any delve client can connect to the server and
debug the plugin.

## Preview documentation

We use [Hugo](https://gohugo.io) to build our documentation site, and it is hosted on
[Netlify](https://netlify.com). You don't have to install Hugo locally because the
preview happens inside a docker container.

1. Run `mage DocsPreview` to start serving the docs. It will watch the file
system for changes.
1. Our make rule should open <http://localhost:1313/docs> to preview the
site/docs.

We welcome your contribution to improve our documentation, and we hope it is an
easy process! ❤️

## Write a blog post

Thank you for writing a post for our blog! 🙇‍♀️ Here's what you need to do to create
a new blog post and then preview it:

1. Go to /docs/content/blog and create a new file. Whatever you name the file
    will be the last part of the URL. For example a file named
    "porter-collaboration.md" will be located at
    <https://getporter.org/blog/porter-collaboration/>.
    
1. At the top of the file copy and paste the frontmatter template below. The
    frontmatter is YAML that instucts the blogging software, Hugo, how to render the
    blog post.
    
    ```yaml
   ---
   title: "Title of Your Blog Post in Titlecase"
   description: "SEO description of your post, displayed in search engine results."
   date: "2020-07-28"
   authorname: "Your Name"
   author: "@yourhandle" #Not used to link to github/twitter, but informally that's what people put here
   authorlink: "https://link/to/your/website" # link to your personal website, github, social media...
   authorimage: "https://link/to/your/profile/picture" # Optional, https://github.com/yourhandle.png works great
   tags: [] # Optional, look at other pages and pick tags that are already in use, e.g. ["mixins"]
   ---
   ```

1. [Preview](#preview-documentation) the website and click "Blog" at the top 
    right to find your blog post.

1. When you create a pull request, look at the checks run by the pull request,
    and click "Details" on the **netlify/porter/deploy-preview** one to see a live
    preview of your pull request.
    
Our pull request preview and the live site will not show posts with a date in
the future. If you don't see your post, change the date to today's date.

## View a trace of a Porter command

Porter can send trace data about the commands run to an OpenTelemetry backend.
It can be very helpful when figuring out why a command failed because you can see the values of variables and stack traces.

In development, you can use the [otel-jaeger bundle] to set up a development instance of Jaeger, which gives you a nice website to see each command run.

```
porter install --reference ghcr.io/getporter/examples/otel-jaeger:v0.1.0 --allow-docker-host-access
```

Then to turn on tracing in Porter, set the following environment variables.
This tells Porter to turn on tracing, and connect to OpenTelemetry server that you just installed.

**Posix**
```bash
export PORTER_TELEMETRY_ENABLED="true"
export OTEL_EXPORTER_OTLP_PROTOCOL="grpc"
export OTEL_EXPORTER_OTLP_INSECURE="true"
```

**Powershell**
```powershell
$env:PORTER_TELEMETRY_ENABLED="true"
$env:OTEL_EXPORTER_OTLP_PROTOCOL="grpc"
$env:OTEL_EXPORTER_OTLP_INSECURE="true"
```

Next run a Porter command to generate some trace data, such as `porter list`.
Then go to the Jaeger website to see your data: http://localhost:16686.
On the Jaeger dashboard, select "porter" from the service drop down, and click "Find Traces".

The smoke and integration tests will run with telemetry enabled when the PORTER_TEST_TELEMETRY_ENABLED environment variable is true.

[otel-jaeger bundle]: https://getporter.org/examples/src/otel-jaeger

## Command Documentation

Our commands are documented at <https://getporter.org/cli> and that documentation is
generated by our CLI. You should regenerate that documentation when you change
any files in **cmd/porter** by running `mage DocsGen` which is run every time
you run `mage build`.

## Work on the Porter Operator

Instructions for building the Porter Operator from source are located in its repository: https://github.com/getporter/operator.
Sometimes you may need to make changes to Porter and work on the Operator at the same time.
Here's how to build porter so that you can use it locally:

1. You must be on a feature branch. Not release/v1 or main. This matters because it affects the generated
   docker image tag.
1. Deploy the operator to a KinD cluster by running `mage deploy` from inside the operator repository.
   That cluster has a local registry running that you can publish to, and it will pull images from it, 
   running on localhost:5000.
1. Run the following command from the porter repository to build the Porter Agent image, and publish it
   to the test cluster's registry. `mage XBuildAll LocalPorterAgentBuild`.
1. Edit your AgentConfig in the Porter Operator and set it to use your local build of the porter-agent.

```yaml
apiVersion: porter.sh/v1
kind: AgentConfig
metadata:
  name: porter
  namespace: test # You may need to change this depending on what you are testing
spec:
  porterRepository: localhost:5000/porter-agent
  porterVersion: canary-dev
  serviceAccount: porter-agent
```

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
  like `porter install --reference`.
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
  * **mixin**: enums, functions and interfaces for the mixin framework.
    * **feed**: works with mixin atom feeds
    * **provider**: handles communicating with mixins
  * **porter**: the implementation of the porter commands. Every command in Porter
    has a corresponding function in here.
    * **version**: reusable library used by all the mixins for implementing a mixin
  * **secrets**: used to access porter's secret store through plugins.
  * **storage**: used to access porter's data store through plugins.
  * **templates**: files that need to be compiled into the porter binary with
      version command.
* **scripts**:
  * **install**: Porter [installation](https://getporter.org/install) scripts
  * **setup-dco**: Set up automatic DCO signoff for the developer environment
* **tests** have Go-based integration tests.

## Logging

**Print to the `Out` property for informational messages and send debug messages to the `Err` property.**

Example:

```golang
fmt.Fprintln(p.Out, "Initiating battlestar protocol")
fmt.Fprintln(p.Err, "DEBUG: loading plans from r2d2...")
```

Most of the structs in Porter have an embedded
`get.porter.sh/porter/pkg/context.Context` struct. This has both `Out` and
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

## Breaking Changes

Some changes in Porter break our compatibility with previous versions of Porter.
When that happens, we need to release that change with a new major version number to indicate to users that it contains breaking changes.
When you realize that you may need to make a breaking change, discuss it with a maintainer on the issue or pull request and we'll come up with a plan for how it should be released.
Here are some examples of breaking changes:

* The schema of porter.yaml changed.
* The schema of Porter's [file formats](https://getporter.org/reference/file-formats) changed.
* The schema of Porter's [config file](https://getporter.org/configuration/#config-file) changed.
* Flags or behavior of a CLI command changed, such as removing a flag or adding a validation that can result in a hard error, preventing the command from running.

All of Porter's documents have a schemaVersion field and when the schema of the document is changed, the version number should be incremented as well in the default set on new documents, the supported schema version constant in the code, and in the documentation for that document.

# Infrastructure

This section includes overviews of infrastructure Porter relies on, mostly intended
for maintainers.

## CDN Setup

See the [CDN Setup Doc][cdn] for details on the services Porter uses to
host and distribute its release binaries.

## Custom Windows CI Agent

Some of our tests need to run on Windows, like the Smoke Tests - Windows stage of our build pipeline.
We use a custom Windows agent registered with Azure Pipelines that we build and maintain ourselves.
See the [Custom Windows CI Agent] documentation for details on how the agent is created and configured.

## Releases

Our [version strategy] explains how we version the project, when you should expect
breaking changes in a release, and the process for the v1 release.

[cdn]: https://getporter.org/src/infra/cdn.md
[version strategy]: https://getporter.org/project/version-strategy/
[Custom Windows CI Agent]: https://getporter.org/src/infra/custom-windows-ci-agent.md
