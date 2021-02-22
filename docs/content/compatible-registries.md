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

| Registry | Compatible |
| -------- | --------------- |
| **Azure Container Registry (ACR)** | **Yes** |
| Artifactory | No |
| **Docker Hub** | **Yes** |
| **DigitalOcean Container Registry** | **Yes** |
| Amazon Elastic Container Registry (ECR) | No |
| **Google Artifact Registry** | **Yes** | 
| Google Cloud Registry (GCR) | No | 
| **GitHub Container Registry (GHCR)** | **Yes** | 
| GitHub Packages | No |
| **Harbor 2** | **Yes** |
| Nexus | No |
| Quay | No |

If you are a registry or user and know that this page is out of date, [please
let us know!](https://github.com/deislabs/porter/issues/new)

[cnab-to-oci]: https://github.com/docker/cnab-to-oci
