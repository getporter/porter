# Philosophy
PRs are most welcome!

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

# Documentation

We use [Hugo](gohugo.io) to build our documentation site, and it is hosted on [Netlify](netlify.com).

1. [Install Hugo](https://gohugo.io/getting-started/installing) using `brew install hugo`, 
`choco install hugo` or `go get -u github.com/gohugoio/hugo`.
1. Run `make docs-preview` to start Hugo. It will watch the file system for changes.
1. Open <http://localhost:1313> to preview the site.

If anyone is interested in contributing changes to our makefile to improve the authoring exerience, such 
as doing this with Docker so that you don't need Hugo installed, it would be a welcome contribution! ❤️