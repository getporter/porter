---
title: Distribute a Mixin
description: How to distribute a mixin for others to install and use
---

Once you have created a mixin, it is time to share it with others so that
they can try it out and use it too. Porter has built-in commands for
managing mixins. All you need to do is get your mixin ready, and publish
them to a file server:

* [Prepare](#prepare)
* [Publish](#publish)
* [Install](#install)

You can also search for other mixins and add your published mixin to Porter's
directory:

* [Search](#search)
* [Broadcast](#broadcast)

## Prepare

Your mixin should be compatible with the following architectures and operating
systems:

* GOOS: windows, linux, darwin
* GOARCH: amd64

If you are creating your mixin in Go, you may find the [Makefile][mk] that we use
for our Porter mixins helpful as a starting point.

## Publish

Porter expects mixins to be published with a specific naming convention:

* client: `VERSION/MIXIN-GOOS-GOARCH[FILE-EXT]`
* runtime: `VERSION/MIXIN-linux-amd64`

\* Note: Porter uses `GOOS` and `GOARCH` which are terms from the Go programming
language, because it is written in Go. You must use the same terms, for example
`darwin` and not macos, in order to Porter to recognize your mixin properly.

Here is an example from the exec mixin:

```
base url/
└── v0.4.0-ralpha.1+dubonnet
    ├── exec-darwin-amd64
    ├── exec-linux-amd64
    └── exec-windows-amd64.exe
```

If you are distributing your mixing via Github releases, upload just the mixin
executables as artifacts to a release named after the mixin version, and it will
match exactly what Porter expects. Then provide the following URL to your users,
`https://github.com/org/project/releases/download`.

## Install

When porter installs a mixin, it builds a url from the command-line arguments:

```
porter mixin install NAME --version VERSION --url URL
```

* client url: `URL/VERSION/NAME-GOOS-GOARCH[FILE_EXT]`
* runtime url: `URL/VERSION/NAME-linux-amd64`

When `--version` is not specified, it is defaulted to `latest` which should
represent the most recent version of the mixin.

You may also choose to publish `canary` versions of the mixin, which are
unpublished builds from the master branch. The official Porter mixins follow
this pattern. If you have other published tagged builds of your mixin, porter
can handle installing them as well.

## Search

Porter maintains a list of remote mixins that users can search via
`porter mixins search [NAME]`. If no name is supplied, the full listing will be
returned.  See the help menu for all command options: `porter mixins search -h`.

For example, here we search for an `az` mixin:

```console
$ porter mixin search az
Name   Description                    Author           URL                                     URL Type
az     A mixin for using the az cli   Porter Authors   https://cdn.porter.sh/mixins/atom.xml   Atom Feed
```

## Broadcast

To add your mixin to the list, add an entry to the mixin directory
[list](https://github.com/deislabs/porter/blob/master/pkg/mixin/directory/index.json)
with all the pertinent informational fields filled out.

For instance, a new entry might look like:

```json
  {
    "name": "mymixin",
    "author": "My Name",
    "description": "A mixin for doing great things",
    "URL": "https://github.com/org/project/releases/download",
  },
```

The `URL` field should either be an Atom Feed URL (for example, Porter uses
the following for its stable mixins: `https://cdn.porter.sh/mixins/atom.xml`) or
a download URL (like the GitHub download URL shown above:
`https://github.com/org/project/releases/download`)

With this change pushed to a branch on your fork of this repo, you're
now ready to open up a [Pull Request](https://github.com/deislabs/porter/pulls).
After the changes are approved, merged and included in the next release, your
mixin will start to show up in the official listing!

[mk]: https://github.com/deislabs/porter/blob/master/mixin.mk
