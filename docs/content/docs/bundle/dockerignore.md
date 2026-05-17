---
title: .dockerignore
description: Using .dockerignore to control which files are copied into your bundle image
---

When Porter builds a bundle, it copies files from your bundle directory into the
installer image.
A `.dockerignore` file in your bundle directory tells Docker which files and
directories to exclude from that copy, keeping the image smaller and build times
shorter.

# Default .dockerignore

When you run `porter create`, Porter generates a `.dockerignore` for you:

```
# See https://docs.docker.com/engine/reference/builder/#dockerignore-file
# Put files here that you don't want copied into your bundle image
.gitignore
template.Dockerfile
```

Both entries are excluded by default because they are only needed on the
developer's machine and have no use inside the installer image.

# Why It Matters

For non-trivial bundles the bundle directory often contains files that are only
needed during development — test fixtures, local configuration, documentation,
or tool caches such as `node_modules`.
Including them inflates the installer image and slows every `porter build` run.
Adding them to `.dockerignore` reduces image size and speeds up builds without
affecting runtime behaviour.

# Example

```
# Development-only files
.gitignore
template.Dockerfile

# Test fixtures and local configuration
tests/
*.local.yaml

# Tool caches
node_modules/

# Local secrets — never ship these
.env
*.pem
```

# Syntax

`.dockerignore` uses the same pattern syntax as `.gitignore`.
See the [Docker .dockerignore reference](https://docs.docker.com/reference/dockerfile/#dockerignore-file)
for the full specification.

# Build Modes

Porter respects `.dockerignore` in both the default and the
[optimized-bundle-build](/docs/bundle/custom-dockerfile/#optimized-build-context-experimental)
build modes.
See [Custom Dockerfile](/docs/bundle/custom-dockerfile/) for details on how
files are copied into the image and how the two build modes differ.
