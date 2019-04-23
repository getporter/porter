---
title: FAQ
description: Frequently Asked Questions
---

* [What is CNAB?](https://cnab.io)
* [Does Porter Replace Duffle?](porter-or-duffle.md)
* [How does your release naming scheme work?](#how-does-your-release-naming-scheme-work)

## How does your release naming scheme work?

Porter's version numbers may look a little funny:

```
v0.4.0-ralpha.1+dubonnet
```

Porter follows [semver](semver.org) but it can help to explain how we apply meaning to
each part:

* `0.4.0` - This is the main version number, Major.Minor.Patch and follow the tenants of semver.
* `ralpha` - Ralph is our PM, so it's "ralpha" instead of "alpha". This indicates that it's not stable, just like Ralph. 😉
* `1` - This means that it's the first build of `v0.4.0-ralpha`. If we need to fix things
on our build system side, but it doesn't affect the shipped code, we may bump that number.
* `dubonnet` - This is the name of the release. We use this name to coordinate across other
CNAB products, like duffle and the VS Code extensions. When another product also has the
same release name, it means that they will work together. The entire version number doesn't
have to match, just the release name. This allows one of the products to ship on their own
release cadence, and still guarantee that a version will work with another product.
