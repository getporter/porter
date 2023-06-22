---
title: Version Strategy
description: How Porter is versioned
---

Porter uses GitHub flow for the most part:

* Branches are created from **main** and merged back into main.
* Changes to main are made available through our **canary** builds.
  These represent the most recent changes to main, and are not stable.
* Every week or so the maintainers determine if there are changes we want to release, and once canary is stable, we tag main with a semantic version.
* You can get the most recent stable release by using the **latest** tag, which is a pointer to the most recent tagged release of the main branch.

## Version Schema

Porter's version numbers adhere [semver v2], of `MAJOR.MINOR.PATCH-PRERELEASE.PRERELEASE_NUMBER`.
The version tracks changes to Porter's configuration files, commands, data format, or behavior.
Porter's library is not yet stable and changes to the underlying Porter code, including breaking changes to downstream consumers, is not encoded in Porter's version number.

* **MAJOR** - Indicates a breaking change. Until we reach v1, there isn't any indication in the version number to indicate breaking changes.
* **MINOR** - Indicates a new feature. Until we reach v1, breaking changes can be included in a minor release.
* **PATCH** - Indicates a bug fix.
* **PRERELEASE** - The name of the prerelease phase, such as alpha, beta or rc (release candidate).
* **PRERELEASE_NUMBER** - The number of releases in the specified prerelease phase.
  For example, v1.0.0-alpha.2 is the second v1 alpha release.

[semver v2]: https://semver.org/spec/v2.0.0.html
