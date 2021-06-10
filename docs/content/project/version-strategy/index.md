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

## v1 Release Plan

Porter v1 will include a number of breaking changes that we are grouping together to minimize disruption to our users.

![drawing of v1 release plan](v1-branch-strategy.jpg)

* Pull requests with v1 work items are branched from the release/v1 branch and merged into the release/v1 branch.
* High severity bug fixes for the stable release of Porter are made against the main branch.
* The goal is to only release patches for the stable version of Porter until v1.0.0 is released.
* Commits to main are cherry-picked or merged into the release/v1 branch.
* Periodic releases of the release/v1 branch will be made, e.g. v1.0.0-alpha.1, v1.0.0-alpha.2, v1.0.0-beta.1, v1.0.0-rc.1.
  The final release from the v1 branch will be v1.0.0.
* The release/v1 branch will be merged into main, and then the v1.0.0 release is cut.
* The latest and canary builds continue to be based on builds of the main branch only.
  We may provide v1-latest and v1-canary builds at a later date.

[semver v2]: https://semver.org/spec/v2.0.0.html
