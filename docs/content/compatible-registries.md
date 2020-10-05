---
title: Compatible Registries
description: Understanding which OCI registries work with CNAB
---

Cloud Native Application Bundles are very new, and support for storing anything
other than container images in a registry is inconsistent. The community has
tested a bunch of registries for compatibility with CNAB and so far here is what
we have found.

There is an explicit verification using Porter because we use specific libraries,
such as [cnab-to-oci], and this helps us communicate confidently that we've tested
out a particular registry and know that it will work for you.

| Registry | CNAB Compatible | Porter Verified |
| -------- | --------------- | ------------- |
| ACR | Yes | Yes |
| Artifactory | No |  |
| Docker Hub | Yes | Yes |
| Digital Ocean | Yes | Yes |
| ECR | No |  |
| GCR | Yes |
| GitHub Packages | Yes | Yes |
| Harbor 2 | yes | yes |
| Nexus | No |  |
| Quay | No | No |

If you are a registry or user and know that this page is out of date, [please
let us know!](https://github.com/deislabs/porter/issues/new)_

[cnab-to-oci]: https://github.com/docker/cnab-to-oci
