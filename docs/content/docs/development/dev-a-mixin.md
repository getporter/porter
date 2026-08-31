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

## Naming a Mixin

Mixin names are used directly in install URLs and file paths, so Porter enforces a naming **rule**:

* Names must be lowercase and may only contain letters, numbers, dashes, and underscores (`[a-z0-9_-]+`).

We also ask that you follow these **conventions** when naming a mixin, though they aren't enforced by Porter:

* Don't include "porter" in the mixin's name.
* Follow our [Code of Conduct][conduct]; no profanity or offensive language.

This same rule applies to plugin names as well.

[conduct]: /src/CODE_OF_CONDUCT.md

## See Also
