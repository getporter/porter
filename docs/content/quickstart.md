---
title: QuickStart Guide
descriptions: Get started using Porter
---

## Getting Porter

First make sure Porter is installed.
Please see the [installation instructions](/install/) for more info.

## Create a new bundle
Use the `porter create` command to start a new project:

```
mkdir -p my-bundle/ && cd my-bundle/
porter create
```

This will create a file called **porter.yaml** which contains the configuration
for your bundle. Modify and customize this file for your application's needs.

This is a snippet of the **porter.yaml** file made by `porter create` to give you
an idea of what it will look like. This bundle is using a helper script, **helpers.sh**,
so that we aren't trying to embed bash inside yaml.

```yaml
name: porter-hello
version: 0.1.0
description: "An example Porter configuration"
registry: getporter

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

## Build the bundle

The `porter build` command will generate the bundle:

```
porter build
```

## Install the bundle

You can then use `porter install` to install your bundle:

```
porter install
```

If you wish to uninstall the bundle, you can use `porter uninstall`:

```
porter uninstall
```

## Publish the bundle

When you are ready to share your bundle, the next step is publishing it to an
OCI registry such as Docker Hub or Quay.

You must authenticate with `docker login` before publishing the bundle. Make
sure that the `registry` listed in your `porter.yaml` is a reference to which the
currently logged in user has write permission.

```yaml
registry: myregistry
```

Now run `porter publish` and porter will push the invocation image and bundle to
the locations specified in the **porter.yaml** file:

```
porter publish
```

## Install from the registry

Now that your bundle is in a registry, anyone can use a [CNAB-compliant
tool][tools], not just Porter, to install the bundle. 

Previously when we use
`porter install` when we were in the same directory as a porter bundle, we
didn't specify an installation name to create, so Porter defaulted the
installation to the name of the bundle. This time we will explicitly name the
installation "demo".

```
porter install demo --reference getporter/porter-hello:v0.1.0
```

[tools]: https://cnab.io/community-projects/#tools

## Cleanup

```
porter uninstall demo
```