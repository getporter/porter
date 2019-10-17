---
title: Copying Bundles
description: How to copy a bundle from one registry to another
---

Porter allows you to copy a bundle, and all associated images, from one registry to another. This is useful when you have restrictions on where you can pull Docker images from or would otherwise like to have control over assets you will be using. Any operation on the copied bundle will utilize these copied images as well. If you'd like to rename something within a registry, Porter also allows you to copy from one bundle tag to another bundle tag inside the same registry.

## Copy A Bundle From One Registry to Another

The first way that you can use the `copy` command is to copy a bundle to a new registry. This command will result in the `source` bundle being copied to the specified registry. The `name:tag` of the source bundle will be used to name the new bundle copy. For example:

```
$ porter copy --source jeremyrickard/porter-do-bundle:v0.4.6 --destination jrrporter.azurecr.io
Beginning bundle copy to jrrporter.azurecr.io/porter-do-bundle:v0.4.6. This may take some time.
Starting to copy image jeremyrickard/porter-do:v0.4.6...
Completed image jeremyrickard/porter-do:v0.4.6 copy
Starting to copy image jeremyrickard/spring-music@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f...
Completed image jeremyrickard/spring-music@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f copy
Bundle tag jrrporter.azurecr.io/porter-do-bundle:v0.4.6 pushed successfully, with digest "sha256:38d08d6e1ecc97dbf22c630309c2ad37e5af6c092b02826aa4285ec24b4765b9"
```

In this case, we copied `jeremyrickard/porter-do-bundle:v0.4.6` to the new registry `jrrporter.azurecr.io`, resulting in the new bundle reference `jrrporter.azurecr.io/porter-do-bundle:v0.4.6`.

## Copy A Bundle From One Registry to Another With Specific Tag

In addition to specifying only the new registry, you can also specify a new tagged reference for the bundle:

```
$ porter copy --source jeremyrickard/porter-do-bundle:v0.4.6 --destination jrrporter.azurecr.io/do-bundle:v0.1.0
Beginning bundle copy to jrrporter.azurecr.io/porter-do-bundle:v0.4.6. This may take some time.
Starting to copy image jeremyrickard/porter-do:v0.4.6...
Completed image jeremyrickard/porter-do:v0.4.6 copy
Starting to copy image jeremyrickard/spring-music@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f...
Completed image jeremyrickard/spring-music@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f copy
Bundle tag jrrporter.azurecr.io/porter-do-bundle:v0.4.6 pushed successfully, with digest "sha256:38d08d6e1ecc97dbf22c630309c2ad37e5af6c092b02826aa4285ec24b4765b9"
```

This results in `jeremyrickard/porter-do-bundle:v0.4.6` being copied to `jrrporter.azurecr.io/do-bundle:v0.1.0`.
