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

# Logging

**Print to the `Out` property for informational messages and send debug messages to the `Err` property.**

Example:

```golang
fmt.Fprintln(p.Out, "Initiating battlestar protocol"
fmt.Fprintln(p.Err, "DEBUG: loading plans from r2d2...")
```

Most of the structs in Porter have an embedded `github.com/deislabs/porter/pkg/context.Context` struct. This has both 
`Out` and `Err` which represent stdout and stderr respectively. You should log to those instead of directly to 
stdout/stderr because that is how we capture output in our unit tests. That means use `fmt.Fprint*` instead of 
`fmt.Print*` so that you can pass in `Out` or `Err`.

Some of our commands are designed to be consumed by another tool and intermixing debug lines and the command output 
would make the resulting output unusable. For example, `porter schema` outputs a json schema
and if log lines were sent to stdout as well, then the resulting json schema would be unparsable. This is why we send
regular command output to `Out` and debug information to `Err`. It allows us to then run the command and see the debug 
output separately, like so `porter schema --debug 2> err.log`.

# Documentation

We use [Hugo](gohugo.io) to build our documentation site, and it is hosted on [Netlify](netlify.com).

1. [Install Hugo](https://gohugo.io/getting-started/installing) using `brew install hugo`, 
`choco install hugo` or `go get -u github.com/gohugoio/hugo`.
1. Run `make docs-preview` to start Hugo. It will watch the file system for changes.
1. Open <http://localhost:1313> to preview the site.

If anyone is interested in contributing changes to our makefile to improve the authoring experience, such 
as doing this with Docker so that you don't need Hugo installed, it would be a welcome contribution! ❤️

[good-first-issue]: https://waffle.io/deislabs/porter?search=backlog&label=good%20first%20issue
[help-wanted]: https://waffle.io/deislabs/porter?search=backlog&label=help%20wanted

# Cutting a Release

🧀💨

Our CI system watches for tags, and when a tag is pushed, it executes the
publish target in the Makefile. When you are asked to cut a new release,
here is the process:

1. Figure out the correct version number, we follow [semver](semver.org) and
    have a funny [release naming scheme][release-name]:
    * Bump the major segment if there are any breaking changes.
    * Bump the minor segment if there are new features only.
    * Bump the patch segment if there are bug fixes only.
    * Bump the build segment (version-prerelease.BUILDTAG+releasename) if you only
      fixed something in the build, but the final binaries are the same.
1. Figure out if the release name (version-prerelease.buildtag+RELEASENAME) should
    change.
    
    * Keep the release name the same if it is just a build tag or patch bump.
    * It is a new release name for major and minor bumps.
    
    If you need a new release name, it must be conversation with the team.
    [Release naming scheme][release-name] explains the meaning behind the
    release names.
1. Ensure that the master CI build is passing, then make the tag and push it.

    ```
    git checkout master
    git pull
    git tag VERSION -a -m ""
    git push --tags
    ```

1. Generate some release notes and put them into the release on GitHub.
    The following command gives you a list of all the merged pull requests:

    ```
    git log --oneline OLDVERSION..NEWVERSION  | grep "#" > gitlog.txt
    ```

    You need to go through that and make a bulleted list of features
    and fixes with the PR titles and links to the PR. If you come up with an
    easier way of doing this, please submit a PR to update these instructions. 😅

    ```
    # Features
    * PR TITLE (#PR NUMBER)

    # Fixes
    * PR TITLE (#PR NUMBER)

    # Install or Upgrade
    Run (or re-run) the installation from https://porter.sh/install to get the 
    latest version of porter.
    ```
1. Name the release after the version.
    

[release-name]: https://porter.sh/faq/#how-does-your-release-naming-scheme-work