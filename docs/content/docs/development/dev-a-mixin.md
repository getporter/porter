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

## See Also
