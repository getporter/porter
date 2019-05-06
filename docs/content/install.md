---
title: Install Porter
description: Installing the Porter client
---

We have a few release types available for you to use:

* **canary**: tip of master
* **vX.Y.Z**: official release
* **latest**: most recent release

You can change the URLs below replacing `latest` with `canary` or a version number
like `v0.1.0-ralpha.1+aviation`.

## MacOS
```
curl https://deislabs.blob.core.windows.net/porter/latest/install-mac.sh | bash
```

## Linux
```
curl https://deislabs.blob.core.windows.net/porter/latest/install-linux.sh | bash
```

## Windows
```
iwr "https://deislabs.blob.core.windows.net/porter/latest/install-windows.ps1" -UseBasicParsing | iex
```

# Mixins

We have a number of [mixins](/mixins) to help you get started. The stable ones are installed
by default by the script:

* exec
* kubernetes
* helm
* azure

You can install a new version of a mixin, or install a mixin that someone else made
using the `porter mixin install` command built into porter.

```console
$ porter mixin install terraform --feed-url https://cdn.deislabs.io/porter/atom.xml
installed terraform mixin
terraform mixin v0.1.0-ralpha.1+elderflowerspritz (edf8778)
```

All of the DeisLabs created mixes are published to the same feed: `https://cdn.deislabs.io/porter/atom.xml`.