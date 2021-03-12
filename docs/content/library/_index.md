---
title: "Porter Go Library"
description: "How to use Porter's Go libraries to programmatically automate Porter"
---

Porter's CLI is built upon a [public Go
library](https://pkg.go.dev/get.porter.sh/porter/pkg/porter) that is available
to anyone who would like to automate Porter programmatically, or access a useful
bit of functionality that isn't exposed perfectly through the CLI.

🚨 Porter does not guarantee backwards compatibility in its library. From release to release
we may make breaking changes to support new features or fix bugs in Porter. Especially before
we reach v1.0.0, as more refactoring **is** going to happen.

Every Porter command is backed by a single function in the
`get.porter.sh/pkg/porter` package that accepts a struct defining the flags and
arguments specified at the command line. You should ALWAYS call `opts.Validate`
when it is defined because that contains defaulting logic.

We recommend using the `porter.Porter` struct, which is created by
`porter.New()` when automating Porter. There are more functions and packages
exposed, but those are much more likely to change over time.

<script src="https://gist-it.appspot.com/https://github.com/getporter/porter/blob/main/pkg/porter/examples/install_example_test.go"></script>

If you need to set stdin/stdout/stderr, you can set `Porter.Out`. The example below demonstrates how to capture stdout.

<script src="https://gist-it.appspot.com/https://github.com/getporter/porter/blob/main/pkg/porter/examples/capture_output_example_test.go"></script>
