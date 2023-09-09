---
title: Bundles
description: Introduction to Bundles
weight: 1
---

A bundle, as defined by the CNAB specification, is a standard packaging format for multi-component distributed applications. It allows packages to target different runtimes and architectures. It empowers application distributors to package applications for deployment on a wide variety of cloud platforms, providers, and services. Furthermore, it provides necessary capabilities for delivering multi-container applications in disconnected (airgapped) environments.

## Bundle Operations

A bundle created by Porter can then be an OCI artifact which includes a docker image with the necessary files and development tools along with the bundle's metadata and any other docker images used by the bundle. Once a bundle is built, it can be distributed using Docker / OCI registries.
This allows you to use existing tools and infrastructure to share your application with other teams, customers, and end-users.

Because the bundle contains everything you need to deploy, included referenced images, you can even [move a bundle into a disconnected or airgapped environment](/administrators/airgap/). When written with airgap deployments in mind, a bundle can be deployed anywhere without requiring access to the original network or the internet.

Once a bundle is installed, is tracked as an "installation" of the bundle.
An installation is a record of which bundle was installed, the parameters used to customize the installation, previous runs of the bundle, the current version of the bundle, along with other useful metadata.

## Example Bundle porter.yaml

```yaml
schemaVersion: 1.0.0-alpha.1
name: examples/porter-hello
version: 0.2.0
description: "An example Porter configuration"
registry: ghcr.io/getporter

parameters:
  - name: name
    type: string
    default: porter
    path: /cnab/app/foo/name.txt
    source:
      output: name

outputs:
  - name: name
    path: /cnab/app/foo/name.txt

mixins:
  - exec

install:
  - exec:
      description: "Install Hello World"
      command: ./helpers.sh
      arguments:
        - install

upgrade:
  - exec:
      description: "World 2.0"
      command: ./helpers.sh
      arguments:
        - upgrade

uninstall:
  - exec:
      description: "Uninstall Hello World"
      command: ./helpers.sh
      arguments:
        - uninstall
```

## Next Steps

- Learn more about Bundles - TODO
- [Introduction to Parameters](/introduction/intro-parameters)
