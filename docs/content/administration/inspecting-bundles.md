---
title: Inspecting Bundles
description: Inspect a bundle to see the images that will be used.
aliases:
- /inspecting-bundles/
- /inspect-bundle/
---

You've found a bundle that you'd like to use, but you'd like to what images will be used after you install the bundle. You can use the `porter inspect` command to see this information. If you'd like to see additional information, like parameters, credentials, and outputs, see the [explain](/examine-bundles) command.

When a bundle is published, the images that it will use are copied into the location of the published bundle. This simplifies access control and management of artifacts in the repository. The `inspect` command will show the invocation images, as well as any referenced images, that will be used as a result of performing actions like install nad upgrade. For each image, you will see the image reference that will be used, along with the original image reference that the image was copied from.

```console
$ porter inspect jeremyrickard/porter-do-bundle:v1.0.0
Name: spring-music
Description: Run the Spring Music Service on Kubernetes and Digital Ocean PostgreSQL
Version: 1.0.0

Invocation Images:
Image                                                                                                    Type     Digest                                                                    Original Image
jeremyrickard/porter-do-bundle@sha256:2fb1f0abdd407e72393e40f411ba60e3eaae505161f49f5fd4c801e1528bbc3f   docker   sha256:2fb1f0abdd407e72393e40f411ba60e3eaae505161f49f5fd4c801e1528bbc3f   jeremyrickard/porter-do:v1.0.0

Images:
Name           Type     Image                                                                                                    Digest                                                                    Original Image
spring-music   docker   jeremyrickard/porter-do-bundle@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f   sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f   jeremyrickard/spring-music:v1.0.0
```

With the image information above, you can use existing tooling to pull, inspect and vet the images before you run the bundle. If you copy or archive and then republish a bundle, the image information will reflect the new locations of the images, allowing you to compare between the source and the new bundle as well. This is especially useful when used with bundles that have been re-published from an archive:

```console
$ porter inspect jrrporter.azurecr.io/do-porter-from-archive:1.0.0
Name: spring-music
Description: Run the Spring Music Service on Kubernetes and Digital Ocean PostgreSQL
Version: 1.0.0

Invocation Images:
Image                                                                                                                 Type     Digest                                                                    Original Image
jrrporter.azurecr.io/do-porter-from-archive@sha256:74b8622a8b7f09a6802a3fff166c8d1827c9e78ac4e4b9e71e0de872fa5077be   docker   sha256:74b8622a8b7f09a6802a3fff166c8d1827c9e78ac4e4b9e71e0de872fa5077be   jrrporter.azurecr.io/do-porter-from-archive/porter-do@sha256:74b8622a8b7f09a6802a3fff166c8d1827c9e78ac4e4b9e71e0de872fa5077be

Images:
Name           Type     Image                                                                                                                 Digest                                                                    Original Image
spring-music   docker   jrrporter.azurecr.io/do-porter-from-archive@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f   sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f   jrrporter.azurecr.io/do-porter-from-archive/spring-music@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f
```

Here, we can see the original image reference and our newly created archive reference. We can compare the two in order to ensure that they are indeed the same.

`porter inspect` can be used with a published bundle, as show above, or with a local bundle. The command even works with bundles that were not built with Porter, through the use of the `--cnab-file` flag. You can view the output in tabular form, as above, JSON or YAML. For all the options, run the command `porter inspect --help`.
