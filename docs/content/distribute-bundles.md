---
title: Distribute Bundles
description: Share and distribute your bundles with others
aliases:
- /distributing-bundles/
---

Once you have built a bundle with Porter, the next step is to share the bundle and invocation image so others can use it. Porter uses OCI (Docker) registries to share both CNAB bundle manifest and invocation images.

* [Preparing For Bundle Publishing](#preparing-for-bundle-publishing)
* [Bundle Publish](#bundle-publish)
* [Publish Archived Bundles](#publish-archived-bundles)
* [Image References After Publishing](#image-references-after-publishing)
 
## Preparing For Bundle Publishing

Before you can publish your bundle, you must first run a `porter build` command. This will create the invocation image so it can be pushed to an OCI (Docker) registry along with your CNAB bundle manifest. It's a good idea to work with your bundle and test it locally before you publish it to a registry.

## Bundle Publish

Once you are satisfied with the bundle, the next step is to publish the bundle! Bundle publishing involves pushing both the invocation image and the CNAB bundle manifest to an OCI registry. Porter uses [docker tags](https://docs.docker.com/engine/reference/commandline/tag/) for both invocation images and CNAB bundle manifests. These are defined in your `porter.yaml` file:

```yaml
name: kubernetes
version: 0.1.0
description: "An example Porter bundle using Kubernetes"
registry: getporter
```

This YAML snippet indicates that the bundle will be built and tagged as `getporter/kubernetes:v0.2.0`. This full bundle reference is constructed from the provided `registry`, `name` and `version` fields. We recommend using [semantic versioning](https://semver.org/) for the bundle version.

The generated invocation image name will be auto-derived from the same combination of `registry`, `name` and `version`.  Using the example above, an invocation image with the name of `getporter/kubernetes:porter-HASH` will be built.

Once you have provided values for the fields above, run the `porter build` command one last time to verify that your invocation image can be successfully built.

Next, run the `porter publish` command in order to push the invocation image to the specified repository and to regenerate a CNAB bundle manifest using this newly pushed image. You should see output like the following:

```
$ porter publish
Pushing CNAB invocation image...
The push refers to repository [docker.io/getporter/kubernetes]
0f4d408243ab: Preparing
6573f19b0ef5: Preparing
a6afb08c6a1c: Preparing
3ae25590a14a: Preparing
68146117cef5: Preparing
d163e1b93415: Preparing
e4b20fcc48f4: Preparing
d163e1b93415: Waiting
e4b20fcc48f4: Waiting
68146117cef5: Layer already exists
d163e1b93415: Layer already exists
e4b20fcc48f4: Layer already exists
a6afb08c6a1c: Pushed
0f4d408243ab: Pushed
6573f19b0ef5: Pushed
3ae25590a14a: Pushed
0.1.0: digest: sha256:5e49e21be75fa940d74fbadac02af9cb31cf7f9147c336e8ce1b42a0537aa7f7 size: 1793

Rewriting CNAB bundle.json...
Starting to copy image getporter/kubernetes:porter-1a1c944c8540836ccdb475dd5ea3adf5...
Completed image getporter/kubernetes:porter-1a1c944c8540836ccdb475dd5ea3adf5 copy
Bundle tag docker.io/getporter/kubernetes:v0.2.0 pushed successfully, with digest "sha256:10a41e6d5af73f2cebe4bf6d368bdf5ccc39e641117051d30f88cf0c69e4e456"
```

Note: you can safely ignore the `WARN[0005] reference for unknown type: application/vnd.cnab.config.v1+json` message, if it appears.

When this command is complete, your CNAB bundle manifest and invocation image will have been successfully pushed to the specified OCI registry. It can then be installed with the `porter install` command:

```
$ porter install --reference getporter/kubernetes:v0.2.0 -c kool-kred
installing kubernetes...
executing porter install configuration from /cnab/app/porter.yaml
Install Hello World App
```

The bundle can also be pulled with specified digest:

```
$ porter install --reference getporter/kubernetes@sha256:10a41e6d5af73f2cebe4bf6d368bdf5ccc39e641117051d30f88cf0c69e4e456 -c kool-kred
installing kubernetes...
executing porter install configuration from /cnab/app/porter.yaml
Install Hello World App
```

The latter example ensures immutability for your bundle. After you've initially run `porter publish`, your tagged reference, such as `getporter/kubernetes:v0.2.0` can be updated with subsequent `porter publish` commands. However, the digested version `getporter/kubernetes@sha256:10a41e6d5af73f2cebe4bf6d368bdf5ccc39e641117051d30f88cf0c69e4e456` will not change. If you'd like to publish different version of the bundle, you will need to update minimally the `tag` attribute and optionally the `invocationImage` attribute, before running `porter publish` again.  (Porter will detect the manifest change and automatically run a new bundle and invocation image build prior to publishing.)

## Publish Archived Bundles

The `porter publish` command can also be used to publish an [archived](/archive-bundles/) bundle to a registry. To publish an archived bundle, the publish command is used with the `-a <filename>` and `--reference <repo/name:tag>` flags. For example, to publish a bundle in the `mybunz1.1.tgz` file to `getporter/megabundle:1.1.0`, you would run the following command:

```
porter publish -a mybunz1.1.tgz --reference getporter/megabundle:1.1.0
```

## Image References After Publishing

When a bundle is published, all images [referenced][image-map] by the bundle are
published **inside** the bundle repository. Bundles should not hard-code image
references and instead should use the [images] section and templating so that they
are referencing the published location of the image.

* REGISTRY/ORG/BUNDLE:TAG
  * REGISTRY/ORG/BUNDLE@**INVOCATION_IMAGE_DIGEST**
  * REGISTRY/ORG/BUNDLE@**REFERENCED_IMAGE_1_DIGEST**
  * REGISTRY/ORG/BUNDLE@**REFERENCED_IMAGE_2_DIGEST**

NOTE: Digest refers to the [repository digest][digest] (not the image id).

Consider the following example:

```
name: porter-do-bundle
version: 0.5.0
description: "Run the Spring Music Service on Kubernetes and Digital Ocean PostgreSQL"
registry: jeremyrickard

images:
  spring-music:
      description: "Spring Music Example"
      imageType: "docker"
      repository: "jeremyrickard/spring-music"
      digest: "sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f"
```

When this bundle is published, both the invocation image and the spring-music
image will be copied and stored in the context of the bundle. To see this in
action, you can use the `porter inspect` command to see what images will
actually be used for a given bundle.

```
Name: spring-music
Description: Run the Spring Music Service on Kubernetes and Digital Ocean PostgreSQL
Version: 0.5.0

Invocation Images:
Image                                            Type     Digest            Original Image
jeremyrickard/porter-do-bundle@sha256:74b86...   docker   sha256:74b86...   jeremyrickard/porter-do-bundle-installer:v0.5.0

Images:
Name           Type     Image                                            Digest            Original Image
spring-music   docker   jeremyrickard/porter-do-bundle@sha256:8f113...   sha256:8f113...   jeremyrickard/spring-music@sha256:8f113...
```

Here we can see that the spring-music image was relocated inside the published bundle:

`jeremyrickard/spring-music@sha256:8f113...` → `jeremyrickard/porter-do-bundle@sha256:8f113...`

[images]: /author-bundles/#images

### Archived Bundle Artifacts

Publishing archived bundles will result in intermediate artifacts that are not
used in the final bundle. New repositories are created for each image used by
the bundle on the destination registry. These are side effects of the publishing
process and they are not used by the final bundle.

Using the example above, the following repositories are created and can be
ignored:

* `jeremyrickard/porter-do-bundle/porter-do`
* `jeremyrickard/porter-do-bundle/spring-music`

[digest]: https://github.com/opencontainers/image-spec/blob/master/descriptor.md#digests
[image-map]: /author-bundles/#images
