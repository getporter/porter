---
title: QuickStart Guide
descriptions: Get started building bundles with Porter
---

## Pre-requisites

Docker is currently a prerequisite for using Porter. Docker is used to package up the bundle. 

If you do not have Docker installed, go ahead and [get Docker](https://docs.docker.com/get-docker/). 

## Getting Porter

Next, you need Porter. Follow the Porter [installation instructions](/install/).

## Create a new bundle

Use the `porter create` command to start a new project:

```
mkdir -p my-bundle/ && cd my-bundle/
porter create
```

This will create a file called **porter.yaml** which contains the configuration
for your bundle. This will be the file that you modify and customize for your application's needs.

## Examine the Porter YAML configuration

Let's look more closely at the bundle manifest, [porter.yaml](quickstart/porter-yaml). 

## Build the bundle

Next, build your first bundle, with [porter build](quickstart/build-bundle). 

## Bundle actions

Now that you have a bundle, what can you do with it? Let's look more closely at [bundle actions](quickstart/bundle-actions). 

## Next steps

So you've created, built, installed, and uninstalled your first bundle. What is next?