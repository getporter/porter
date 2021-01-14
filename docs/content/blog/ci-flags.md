---
title: "Set Bundle Metadata on Build and Publish"
description: |
    Integrate porter into your CI/CD pipeline with new build and publish flags
    allowing you to set the name, version, registry and more.
date: "2021-01-14"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://twitter.com/carolynvs"
authorimage: "https://github.com/carolynvs.png"
tags: ["release-notes", "ci-cd"]
---

With the [v0.31.0 release of Porter][v0.31.0] you can now quickly set metadata
on your bundle. This corrects confusing terms around the OCI reference of the
bundle (the location of a bundle in a registry).

<p align=center>reference = registry/name:tag</p>

* **Reference**: A bundle reference is the location of a published bundle. For
  example, **ghcr.<span>io</span>/getporter/porter/porter-hello:v0.1.1**.
  Previously this was called bundle tag.
  
  Until we release v1.0, Porter detects when you use the flags with the old
  meanings and fixes it for you. When Porter releases v1.0, that behavior will
  be removed.

* **Registry**: The location prefix of a published bundle. For example, with
  <strong>ghcr.<span>io</span>/getporter/porter</strong>/porter-hello:v0.1.1 the registry is
  ghcr.<span>io</span>/getporter/porter.

* **Tag**: Tag now _only_ means the OCI artifact tag, which is the last part of
  a bundle or image reference after the colon. For example, with
  getporter/porter-hello:<strong>v0.1.1</strong> the tag is v0.1.1. Previously this had two
  definitions: bundle tag and OCI artifact tag.

# Build Workflow

You can change the bundle name and version when building the bundle. Porter
replaces the specified field in porter.yaml in-memory with the new value before
building the bundle. The porter.yaml file is not modified.

## Name

Set the bundle name with `--name` when building the bundle:

```
porter build --name porter-hello
```

## Version

Set the bundle version with `--version` when building the bundle:

```
porter build --version v0.1.1
```

Most bundle pipelines will determine the version on the fly during the build
either from application version or using git tags.

# Publish Workflow

You can change where the bundle is published using the flags below, and Porter
will retag the bundle before pushing it.

## Tag

Set the tag with `--tag` when publishing the bundle:

```
porter publish --tag latest
```

We don't recommend using mutable tags like latest in production, but they are
very useful during development and test.

## Registry

Set the registry with `--registry` when publishing the bundle:

```
porter publish --registry ghcr.io/getporter
```

This makes it easier to publish a bundle to multiple registries, for example
both Docker Hub and GitHub Container Registry, at the same time.

## Reference

Set the entire reference with `--reference` when publishing the bundle:

```
porter publish --reference localhost:5000/testing/hello:dev
```

Using a combination of the build and publish flags works for most CI/CD
scenarios. If for some reason the other flags don't work for your situation, you
can override the entire reference, bypassing Porter's defaulting logic.

---

Hopefully makes it a lot easier to build and publish a bundle in your CI
pipeline. Try it out and let us know how it works for you!

[v0.31.0]: https://github.com/getporter/porter/releases/v0.31.0
