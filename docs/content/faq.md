---
title: FAQ
description: Frequently Asked Questions
---

* [What is CNAB?](#what-is-cnab)
* [Does Porter Replace Duffle?](#does-porter-replace-duffle)
* [Should I use Porter or Duffle?](#should-i-use-porter-or-duffle)
* [How does your release naming scheme work?](#how-does-your-release-naming-scheme-work)

## What is CNAB?

CNAB stands for "Cloud Native Application Bundle". When we say "bundle", that is what
we are referring to. There is a CNAB Specification and you can learn more about
it at [cnab.io](https://cnab.io).

## Does Porter replace Duffle?

  <p align="center"><strong>No, Porter is not a replacement of Duffle.</strong></p>

In short:

> Duffle is the reference implementation of the CNAB specification and is used 
> to quickly vet and demonstrate a working specification.

> Porter supports the CNAB spec and empowers bundle authors to create composable, 
> reusable bundles using familiar tools like Helm, Terraform, and their cloud provider's 
> CLIs. Porter is designed to be the best user experience for working with bundles.

See [Porter or Duffle](/porter-or-duffle) for a comparison of the tools.

## Should I use Porter or Duffle?

If you are contributing to the CNAB specification, we recommend vetting your contributions by
"verification through implementation" on Duffle.

If you are making bundles, may we suggest using Porter?

<p align="center">👩🏽‍✈️ ️️👩🏽‍✈️ 👩🏽‍✈️</p> 

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


## Does CNAB fully implement the CNAB specification?

Porter currently implements much of the CNAB spec, however, as the [CNAB specification](https://github.com/deislabs/cnab-spec) moves toward 1.0, some gaps have emerged. Currently, if you build a bundle with Porter, you'll be able to install it with Porter. There are some gaps with the spec that limit compatibility with other CNAB tooling. See the [CNAB 1.0 Milestone](https://github.com/deislabs/porter/milestone/12) for more information on these gaps.

