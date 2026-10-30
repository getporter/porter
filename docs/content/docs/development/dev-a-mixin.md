---
title: Developing Mixins
description: Creating and Extending Mixins for Porter
weight: 2
---

You can develop a mixin utilizing our template repository as a start at <https://github.com/getporter/skeletor>
or by using `porter mixin create` command as exemplified below:

```bash
porter mixin create mymixin --author "My Name" --username mygithubusername [--dir path/to/mymixin]
```

> **Note:** We're also developing an official [mixin-sdk-go][mixin-sdk-go] library,
> meant to eventually replace skeletor as the way to build Go mixins. It's still a
> work in progress, so skeletor remains the recommended starting point for now. If
> you'd like to try the SDK early, see its [tutorial][mixin-sdk-go-tutorial].

## Naming a Mixin

Mixin names are used directly in install URLs and file paths, so Porter enforces a naming **rule**:

* Names must be lowercase and may only contain letters, numbers, dashes, and underscores (`[a-z0-9_-]+`).

We also ask that you follow these **conventions** when naming a mixin, though they aren't enforced by Porter:

* Don't include "porter" in the mixin's name.
* Follow our [Code of Conduct][conduct]; no profanity or offensive language.

This same rule applies to plugin names as well.

[conduct]: /src/CODE_OF_CONDUCT.md
[mixin-sdk-go]: https://github.com/getporter/mixin-sdk-go
[mixin-sdk-go-tutorial]: https://github.com/getporter/mixin-sdk-go/blob/main/docs/tutorial.md

## Reading the Porter Version at Runtime

When a bundle is built with a released version of Porter, the bundle image sets the
`PORTER_VERSION` environment variable to that version, for example `v1.2.0`. A mixin
can read it while the bundle runs, for example to include the Porter version in the
user agent it sends to a cloud provider:

```go
userAgent := "getporter/aws/" + pkg.Version
if porterVersion := os.Getenv("PORTER_VERSION"); porterVersion != "" {
	userAgent = "getporter/porter/" + porterVersion + " " + userAgent
}
```

Bundles built with a development build of Porter, which has no version, don't set
`PORTER_VERSION`, so treat it as optional.

## See Also
