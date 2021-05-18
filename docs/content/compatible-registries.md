---
title: Compatible Registries
description: Understanding which OCI registries work with CNAB
---

Cloud Native Application Bundles (CNAB) are very new, and support for storing anything
other than container images in a registry is inconsistent. The community has
tested a bunch of registries for compatibility with CNAB and so far here is what
we have found.

There is an explicit verification using Porter because we use specific libraries,
such as [cnab-to-oci], and this helps us communicate confidently that we've tested
out a particular registry and know that it will work for you.

| Registry | Compatible |
| -------- | --------------- |
| **Azure Container Registry (ACR)** | **Yes** |
| Artifactory | No |
| **Docker Hub** | **Yes** |
| **DigitalOcean Container Registry** | **Yes** |
| **Amazon Elastic Container Registry (ECR)** | **[Yes*](#notes)** |
| **Google Artifact Registry** | **Yes** |
| Google Cloud Registry (GCR) | No |
| **GitHub Container Registry (GHCR)** | **Yes** |
| GitHub Packages | No |
| **Harbor 2** | **Yes** |
| Nexus | No |
| Quay | No |

If you test a registry with Porter and find that this page is out of date, [please
let us know](https://github.com/deislabs/porter/issues/new)!

### Notes

* Amazon Elastic Container Registry (ECR) requires that you create the repository for the installer and the bundle before publishing.

[cnab-to-oci]: https://github.com/cnabio/cnab-to-oci
[oci-spec]: https://github.com/opencontainers/distribution-spec/blob/master/spec.md
