---
title: Quickstart Guide
descriptions: Get started using Porter
---

## Getting Porter

First make sure Porter is installed.
Please see the [installation instructions](./install.md) for more info.

## Creating a new installer
Use the `porter create` command to start a new project:
```
mkdir -p my-installer/ && cd my-installer/
porter create
```

This will create a file called `porter.yaml` which contains the configuration for your installer. Modify and customize this file for your application's needs.

Here is a very basic `porter.yaml` example:
```
name: my-installer
version: 0.1.0
description: "this application is extremely important"
invocationImage: my-dockerhub-user/my-installer:latest
mixins:
  - exec
install:
  - description: "Install Hello World"
    exec:
      command: bash
      arguments:
        - -c
        - echo Hello World
uninstall:
  - description: "Uninstall Hello World"
    exec:
      command: bash
      arguments:
        - -c
        - echo Goodbye World
```

## Building a CNAB bundle

The `porter build` command will create a [CNAB-compliant](https://github.com/deislabs/cnab-spec/blob/master/101-bundle-json.md) `bundle.json`, as well as build and push the associated invocation image:
```
porter build
```

Note: Make sure that the `invocationImage` listed in you `porter.yaml`  is a reference that you are able to `docker push` to and that your end-users are able to `docker pull` from.

## Running your installer using Duffle

_Wondering the differences between Duffle and Porter? Please see [this page](./porter-or-duffle.md)._

[Duffle](https://duffle.sh/) is an open-source tool that allows you to install and manage CNAB bundles.

The file `duffle.json` is required by Duffle. Since both Porter and Duffle are CNAB-complaint, we can simply reuse the `bundle.json` created during `porter build` and run `duffle build` to save it to the local store:
```
cp bundle.json duffle.json
duffle build .
```

You can view all bundles in the local store with `duffle bundle list`.

Afterwords, use `duffle install` to run your installer ("demo" is the unique installation name):
```
duffle install demo my-installer:0.1.0
```

The `duffle list` command can be used to show all installed applications.

If you wish to uninstall the application, you can use `duffle uninstall`:
```
duffle uninstall demo
```