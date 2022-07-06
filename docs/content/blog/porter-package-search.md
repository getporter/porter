---
title: "Searching for Mixins and Plugins with Porter"
description: "Discover mixins and plugins available for Porter and add yours to the list."
date: "2020-03-02"
authorname: "Vaughn Dice"
author: "@vdice"
authorlink: "https://github.com/vdice"
authorimage: "https://github.com/vdice.png"
tags: ["plugins", "mixins"]
---

Two of Porter's great strengths lie in composability and extensibility.
[Mixins][mixins] represent the building blocks for composing a bundle, offering
functionality that can be utilized during the runtime of a bundle, and
[Plugins][plugins] represent swappable backend storage solutions,
enabling distributed, Day 2 bundle instance management. (Be sure to read
Carolyn Van Slyck's recent post around
[Plugins and Collaboration](https://deislabs.io/posts/porter-collaboration/).)

The default [installation][install] of Porter includes a handful of
useful mixins, including:

  * [exec][exec] for running arbitrary shell commands
  * [helm][helm] for managing Helm charts and releases
  * [kubernetes][kubernetes] for managing Kubernetes manifests and cluster operations
  
However, there are many more available for use with Porter. As the list grows
and source code is distributed across many different GitHub orgs/users, we
need a way to efficiently search for available mixins. In addition, a
mechanism for mixin developers to add their own offerings to this searchable
directory is necessary.

We created the [Porter Packages][porter-packages] repository to hold these
directories, one for mixins and one for plugins. Adding an entry to the list
is as simple as supplying a handful of details regarding the package in JSON
form: the name of the package, a description of what it does, a URL where it
can be installed from, etc.

The Porter CLI then references the appropriate list when a user searches,
say, for a Terraform mixin:

```console
$ porter mixin search Terraform
Name        Description                           Author           URL                                     URL Type
terraform   A mixin for using the terraform cli   Porter Authors   https://cdn.porter.sh/mixins/atom.xml   Atom Feed
```


We can then install this mixin via the provided URL, which in this case is of
the Atom feed type:

```console
$ porter mixin install terraform --feed-url https://cdn.porter.sh/mixins/atom.xml
installed terraform mixin v0.5.1-beta.1 (597a442)
```


The [Terraform mixin](/mixins/terraform) is now available for use in our next Porter bundle.
To peruse the full list of mixins, simply issue `porter mixin search` without
any query.

Similarly, `porter plugin search` offers a way to discover plugins for Porter.
Plugins offer storage solutions for bundle instance and credentials using the
cloud of your choosing.  We're just getting started and hope that in the near
future, parameter storage will also be included.

Thank you to the community for all of the fantastic package contributions thus
far! We're excited to see Porter's ecosystem grow with each new mixin or
plugin added to the list.

😎 Have a mixin or plugin not already included in either list? Please visit the
[Porter Packages][porter-packages] GitHub repository and follow the
instructions on how to add new entries to the lists.

🎉 Interested in developing a new mixin? Check out the
[Mixin Development Guide](/mixin-dev-guide/) to get started.
We hope to craft a similar guide for plugin development, but in the meantime,
check out the code for the `azure` plugin via the
[Porter Azure Plugins](https://github.com/deislabs/porter-azure-plugins) repo.

[mixins]: /mixins/
[plugins]: /plugins/
[install]: /install/
[exec]: /mixins/exec/
[helm]: /mixins/helm/
[kubernetes]: /mixins/kubernetes/
[porter-packages]: https://github.com/deislabs/porter-packages
[package-search]: /package-search/
