---
title: Glossary
description: Definitions for terms used in Porter and CNAB documentation
weight: 1
---

## Bundle

A [CNAB](#cnab) packaging format for multi-component distributed applications.
A bundle packages everything needed to install, upgrade, and uninstall an
application, including the tooling and logic required to manage its lifecycle.

See [Introduction to Bundles](/docs/introduction/concepts-and-components/intro-bundles/).

## CNAB

**Cloud Native Application Bundle.** An open specification for packaging and
distributing cloud-native applications and their dependencies. CNAB defines how
bundles are built, published, and executed in a consistent way across different
environments and toolchains.

See the [CNAB specification](https://cnab.io/).

## Config File

The `porter.yaml` manifest at the root of a bundle's source directory. It
defines the bundle's metadata, parameters, credentials, mixins, and the actions
(install, upgrade, uninstall) used to manage the application lifecycle.

## Digest

An immutable SHA256 hash that uniquely identifies a specific image manifest in
a [registry](#registry). Unlike a [tag](#tag), a digest never changes and always
refers to the same content.

Example: `sha256:abc123...`

## Docker Host / Daemon

The Docker service (`dockerd`) running on a machine that manages containers,
images, networks, and volumes. Porter's Docker [mixin](#mixin) communicates with
the Docker daemon to run containers as part of a bundle's actions.

## Installation

Porter's record of a [bundle](#bundle) instance that has been installed into an
environment. An installation tracks the bundle reference, parameter values,
credential bindings, and the current state of the managed application.

See [Introduction to Desired State](/docs/introduction/concepts-and-components/intro-desired-state/).

## Manifest

A JSON document stored in a [registry](#registry) that describes an OCI image,
including its configuration and the list of content-addressable layers that make
up the image filesystem.

## Mixin

An adapter that integrates an existing tool (such as Helm, Terraform, or kubectl)
into a Porter bundle. Mixins provide reusable steps that can be used in a
bundle's install, upgrade, and uninstall actions.

See [Introduction to Mixins](/docs/introduction/concepts-and-components/intro-mixins/).

## Plugin

An extension that provides Porter with alternative backends for storage or
secrets management. Plugins allow Porter to integrate with different
infrastructure (for example, storing installation records in Azure Blob Storage
or retrieving credentials from HashiCorp Vault).

See [Introduction to Plugins](/docs/introduction/concepts-and-components/intro-plugins/).

## Registry

An OCI/Docker-compatible service that stores and distributes images and
artifacts. Bundles are published to and pulled from registries in the same way
as container images.

Examples: Docker Hub, GitHub Container Registry, Azure Container Registry.

## Repository

A named collection of related images within a [registry](#registry). A
repository holds multiple versions of the same artifact, each identified by a
[tag](#tag) or [digest](#digest).

Example: `ghcr.io/getporter/porter` is a repository within the
`ghcr.io` registry.

## Tag

A mutable, human-readable label that points to a specific version of an image
within a [repository](#repository). Because tags can be updated to point to
different content over time, use a [digest](#digest) when you need a stable,
immutable reference.

Example: `ghcr.io/getporter/porter:v1.0.0`
