---
title: "Helm3 Mixin Improvements"
description: "The helm3 mixin now supports more flags and has improved defaults"
date: "2021-09-28"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://carolynvanslyck.com/"
authorimage: "https://github.com/carolynvs.png"
tags: ["mixins", "helm"]
---

If you use Helm in your bundles, then you are going to want the latest helm3 mixin release!
With improved defaults, and new flags supported, the [helm3 v0.1.15 mixin](https://github.com/MChorfa/porter-helm3/releases/tag/v0.1.15) will help make your bundles more reliable.
<!--more-->

If you are still using the [deprecated helm mixin](/blog/helm-mixin-rename/), now's the time to switch over to
[Mohamed Chorfa's Helm3 mixin](https://github.com/MChorfa/porter-helm3).

## New Defaults

The helm3 mixin has changed its default behavior to be more robust and avoid situations where the mixin fails when
an action is retried.

### Upgrade or Install
The upsert setting has been removed because it is now the default behavior for the install and upgrade actions.
Now the mixin always uses `helm upgrade --install` which improves reliability when you need to retry a failed action.

### Create Namespace
The mixin specifies `--create-namespace` by default so that the release namespaces is created automatically when it is not already present.

### Atomic
The `--atomic` flag is now specified by default, failed helm releases are rolled back so that
your release is always in a working state.

## New Settings

The helm3 mixin now supports additional Helm flags:

* The `noHooks` setting prevents hooks from being executed and defaults to false.
* The `skipCrds` setting skips installing CRDs during upgrade. By default, CRDs are installed if not already present.
* The `timeout` setting sets how long Helm will wait for the release to complete successfully before rolling it back (due to the atomic behavior).
* The `debug` setting gives you more insight into why a release failed and was rolled back.

## Removed Settings

* The `upsert` setting has been removed because it is the default behavior.
* The `replace` setting has been removed because it is not recommended for production use, and shouldn't be necessary with the new default behavior.

The Porter project has already upgraded our bundles to use this latest version, and
I recommend upgrading your bundles to take advantage of the improved reliability.
The only hiccup in our upgrade was that the \--atomic behavior made us quickly realize that some of our charts were never actually installing 100% successfully. 🤦‍♀️
If you have charts that don't work well with \--atomic or \--wait, or need additional flags exposed, let us know!
