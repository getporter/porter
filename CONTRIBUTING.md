We have [good first issues][good-first-issue] for new contributors and [help wanted][help-wanted] issues for our other contributors.

* `good first issue` has extra information to help you make your first contribution.
* `help wanted` are issues suitable for someone who isn't a core maintainer.

Maintainers will do our best regularly make new issues for you to solve and then help out as you work on them. 💖
 
# Philosophy
PRs are most welcome!

* If there isn't an issue for your PR, please make an issue first and explain the problem or motivation for
the change you are proposing. When the solution isn't straightforward, for example "Implement missing command X",
then also outline your proposed solution. Your PR will go smoother if the solution is agreed upon before you've
spent a lot of time implementing it.
  * It's OK to submit a PR directly for problems such as misspellings or other things where the motivation/problem is
    unambiguous.
* If you aren't sure about your solution yet, put WIP in the title so that people know to be nice and 
wait for you to finish before commenting.
* Try to keep your PRs to a single task. Please don't tackle multiple things in a single PR if possible. Otherwise, grouping related changes into commits will help us out a bunch when reviewing!
* We encourage "follow-on PRs". If the core of your changes are good, and it won't hurt to do more of
the changes later, we like to merge early, and keep working on it in another PR so that others can build
on top of your work.

# Client

1. `make build`, and the resulting binaries are located in `./bin`.
1. `./bin/porter COMMAND`, such as `./bin/porter build`.

If you would like to install a developer build, run `make install`.
This copies a dev build to `~/.porter` and symlinks it to `/usr/local/bin`.

# Mixins

When you run `make build`, the canary\* build of external mixins are automatically installed into your bin directory
in the root of the repository. If you want to work against a different version of a mixin, then run `make clean build MIXIN_TAG=v1.2.3`.
or use `latest` for the most recent tagged release.

\* canary = most recent successful build of master

# Documentation

We use [Hugo](gohugo.io) to build our documentation site, and it is hosted on [Netlify](netlify.com).

1. [Install Hugo](https://gohugo.io/getting-started/installing) using `brew install hugo`, 
`choco install hugo` or `go get -u github.com/gohugoio/hugo`.
1. Run `make docs-preview` to start Hugo. It will watch the file system for changes.
1. Open <http://localhost:1313> to preview the site.

If anyone is interested in contributing changes to our makefile to improve the authoring exerience, such 
as doing this with Docker so that you don't need Hugo installed, it would be a welcome contribution! ❤️

[good-first-issue]: https://github.com/deislabs/porter/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22+label%3Abacklog+
[help-wanted]: https://github.com/deislabs/porter/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22+label%3Abacklog+