---
title: "Porter v1 on the Horizon"
description: "Porter's v1 roadmap and timeline"
date: "2021-05-11"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://carolynvanslyck.com/"
authorimage: "https://github.com/carolynvs.png"
tags: ["roadmap"]
---

Like a farmer asking "when it's gonna rain?", I am often asked "when you gonna ship v1?"
<!--more-->

The Porter maintainers have tweaked our goals for v1. Mainly to stop chasing feature complete and ship a stable version of what works today.
Our [v1 milestone] focuses on P0 and P1 bugs (informally known as emergency donuts 🚨🍩 and chocolate 🍫 in our issue backlog), usability issues without good workarounds, and batching together breaking changes.

Over the course of Porter's lifetime we have deprecated some configuration and flags, but kept shims in place for those who haven't migrated to the replacements yet.
For example \--tag was replaced by \--reference, but we still support \--tag for now.
In v1 those deprecation shims will be removed, and anyone who hasn't migrated yet will need to update their scripts and porter.yaml files.

In order to isolate you from those changes, we have created a v1 branch where all v1 work will be batched up.
Our [version strategy] outlines our development process while we work on v1 and what to expect in terms of version numbers and stability.
**No breaking changes will be released going forward until we ship the final v1.0.0 release of Porter.**
We will be cutting pre-release versions of Porter that you may test out and prepare for v1.

The canary and latest releases will continue to follow the stable main branch.
We recommend that you stick with latest, or even better a pinned version of Porter, until v1.0.0 is available.
That way you are isolated from breaking changes until you are ready to upgrade.

Don't worry, we are still improving Porter!
New features, such as [buildkit] support, [advanced dependency management], organization with [labels and namespaces], and more are still under development.
They will be released as experimental features that are disabled by default with feature flags.
That way we can keep Porter stable, and give you a way to try out new functionality and provide feedback before they are completed.

This is open source, and we only get things done as quickly as we have consensus and contributors.
While I can't give a concrete timeline, we would love to see a v1 release this summer.
We always welcome new contributors, even if you are new to Go, containers, or even Porter!
If you'd like to help us with v1 or contribute to the larger features in our pipeline, the [new contributor guide] is where to start. 🚀

[v1 milestone]: https://github.com/getporter/porter/milestone/16
[version strategy]: /project/version-strategy/
[buildkit]: https://github.com/getporter/porter/pull/1567
[advanced dependency management]: https://github.com/getporter/proposals/pull/8
[labels and namespaces]: https://github.com/cnabio/cnab-spec/pull/411
[new contributor guide]: /contribute/