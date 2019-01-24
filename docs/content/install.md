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