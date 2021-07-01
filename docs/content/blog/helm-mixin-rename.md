---
title: "Upgrade to the helm3 mixin"
description: "Details on deprecating the helm2 mixin and upgrading to the helm3 mixin"
date: "2021-07-01"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://carolynvanslyck.com/"
authorimage: "https://github.com/carolynvs.png"
tags: ["mixins", "helm"]
---

The time has come... to upgrade to the helm3 mixin.
<!--more-->

Last year Helm v2 was put on ice and all new development and support going forward has been happening on Helm v3.
The helm mixin was developed before Helm v3 and only supports Helm v2, so the time has come for us all to say goodbye to helm mixin and upgrade to the helm3 mixin.

Luckily for us [Mohamed Chorfa maintains the helm3 mixin][helm3] and that is the recommended mixin to use going forward.
If you have previously installing the helm mixin with `porter mixins install helm` the new command is:

```
porter mixin install helm3 --feed-url https://mchorfa.github.io/porter-helm3/atom.xml
```

The old helm mixin has been deprecated and renamed to helm2 to avoid confusion.
The helm mixin will no longer appear in mixin search results, although you can still explicitly install it.

So please upgrade to the helm3 mixin if you haven't already.
And as always let us know if you run into any problems or have feedback!

[helm3]: https://github.com/MChorfa/porter-helm3
