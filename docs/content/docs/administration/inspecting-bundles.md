---
title: Inspecting Bundles
description: Inspect a bundle to see the images that will be used.
weight: 1
aliases:
  - /inspecting-bundles/
  - /inspect-bundle/
---

You've found a bundle that you'd like to use, but you'd like to what images will be used after you install the bundle. You can use the `porter inspect` command to see this information. If you'd like to see additional information, like parameters, credentials, and outputs, see the [explain](/docs/operations/examine-bundles/) command.

When a bundle is published, the images that it will use are copied into the location of the published bundle. This simplifies access control and management of artifacts in the repository. The `inspect` command will show the bundle images, as well as any referenced images, that will be used as a result of performing actions like install nad upgrade. For each image, you will see the image reference that will be used, along with the original image reference that the image was copied from.

```console
$ porter inspect jeremyrickard/porter-do-bundle:v1.0.0
Name: spring-music
Description: Run the Spring Music Service on Kubernetes and Digital Ocean PostgreSQL
Version: 1.0.0

Bundle Images:
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

Bundle Images:
Image                                                                                                                 Type     Digest                                                                    Original Image
jrrporter.azurecr.io/do-porter-from-archive@sha256:74b8622a8b7f09a6802a3fff166c8d1827c9e78ac4e4b9e71e0de872fa5077be   docker   sha256:74b8622a8b7f09a6802a3fff166c8d1827c9e78ac4e4b9e71e0de872fa5077be   jrrporter.azurecr.io/do-porter-from-archive/porter-do@sha256:74b8622a8b7f09a6802a3fff166c8d1827c9e78ac4e4b9e71e0de872fa5077be

Images:
Name           Type     Image                                                                                                                 Digest                                                                    Original Image
spring-music   docker   jrrporter.azurecr.io/do-porter-from-archive@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f   sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f   jrrporter.azurecr.io/do-porter-from-archive/spring-music@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f
```

Here, we can see the original image reference and our newly created archive reference. We can compare the two in order to ensure that they are indeed the same.

### Column Reference

The **Bundle Images** and **Images** tables share the same set of columns:

- **Name** (Images table only): The alias given to the image in the bundle definition, e.g. `spring-music`.
- **Image**: The image reference that will actually be pulled and used. If the bundle was published, copied, or archived, this reflects the relocated reference.
- **Type**: The image type, e.g. `docker`.
- **Digest**: The content digest of the image.
- **Original Image**: The image reference that the current image was copied from. This is only populated when the image has been relocated, such as after a publish, copy, or archive operation. Comparing **Image**/**Digest** against **Original Image** lets you confirm that a republished or archived bundle's images still match their source.

## Inspecting Dependencies

If the bundle declares dependencies, you can view its full dependency tree with the `--show-dependencies` flag:

```console
$ porter inspect ghcr.io/getporter/examples/porter-hello:v0.2.0 --show-dependencies
...
Dependencies:
Alias         Reference                                    Version   Sharing
mysql           ghcr.io/getporter/examples/mysql:v0.1.0     0.1.0
  logging       ghcr.io/getporter/examples/logging:v0.2.0   0.2.0     [shared:logging-group]
```

The **Dependencies** table has the following columns:

- **Alias**: The dependency's alias as declared in the bundle. Nested dependencies are indented to show their position in the dependency tree. A ⚠️ prefix indicates the dependency failed to resolve.
- **Reference**: The bundle reference used for the dependency.
- **Version**: The resolved version of the dependency.
- **Sharing**: Shows `[shared:<group>]` when the dependency participates in a shared instance group, otherwise blank.

If any dependency fails to resolve, a note is printed below the table telling you to rerun with `--output json` or `--output yaml` to see the detailed error message.

Other flags that affect dependency resolution:

- `--max-dependency-depth`: Limits how deep the dependency tree is traversed. Defaults to `10`.
- `--dependencies-version-strategy`: Controls how dependency version ranges are resolved. Allowed values are `exact`, `max-patch`, `max-minor`, and `min`.

`porter inspect` can be used with a published bundle, as show above, or with a local bundle. The command even works with bundles that were not built with Porter, through the use of the `--cnab-file` flag. You can view the output in tabular form, as above, JSON or YAML. For all the options, run the command `porter inspect --help`.
