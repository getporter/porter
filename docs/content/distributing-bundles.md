---
title: Distributing Bundles
description: How to distribute your bundles
---

Once you have built a bundle with Porter, the next step is to share the bundle and invocation image so others can use it. Porter uses OCI (Docker) registries to share both CNAB bundle manifest and invocation images.

## Preparing For Bundle Publishing

Before you can publish your bundle, you must first run a `porter build` command. This will create the invocation image so it can be pushed to an OCI (Docker) registry along with your CNAB bundle manifest. It's a good idea to work with your bundle and test it locally before you publish it to a registry.

## Bundle Publish

Once you are satisfied with the bundle, the next step is to publish the bundle! Bundle publishing involves pushing the both the invocation image and the CNAB bundle manifest to an OCI registry. Porter uses [docker tags](https://docs.docker.com/engine/reference/commandline/tag/) for both invocation images and CNAB bundle manifests. These are defined in your `porter.yaml` file:

```yaml
name: kube-example
version: 0.1.0
description: "An example Porter bundle using Kubernetes"
invocationImage: deislabs/porter-kubernetes:latest
tag: deislabs/porter-kube-bundle:1.0
```

This YAML snippet indicates that the invocation image will be built and tagged as `deislabs/porter-kubernetes:latest`. The first part of this reference, `deislabs` indicates the registry that the invocation image should eventually be published to. The `porter-kubernetes` segment identifies the image, while the `:latest` portion denotes a specific version. Much like the `invocationImage` attribute is used to control the name of resulting Docker invocation image, the `tag` attribute is used to specify the name and location of the resulting CNAB bundle. In both cases, when you are ready to publish your bundle, it would be a good idea to provide specific versions for both of these, such as `v1.0.0`. We recommend using [semantic versioning](https://semver.org/) for both the invocation image and the bundle. We also recommend specifying the same registry for both, in order to simplify access to your bundle and invocation image by end users.

Once you have provided values for these, run the `porter build` command one last time to verify that your invocation image can be successfully built and to ensure that the value you specified in `invocationImage` is correct.

Next, run the `porter publish` command in order to push the invocation image to the specified repository and to regenerate a CNAB bundle manifest using this newly pushed image. You should see output like the following:

```
$ porter publish
Pushing CNAB invocation image...
The push refers to repository [docker.io/deislabs/porter-kubernetes]
c412023fe7ea: Preparing
397a70d3e67f: Preparing
49037d9d1b30: Preparing
c7956a703d1e: Preparing
c581f4ede92d: Preparing
c581f4ede92d: Layer already exists
c7956a703d1e: Layer already exists
49037d9d1b30: Layer already exists
c412023fe7ea: Layer already exists
397a70d3e67f: Layer already exists
latest: digest: sha256:d8aa654f5e60d64f698d79664480500b8de469a22e15dc69806e8172848e17d6 size: 1370

Generating CNAB bundle.json...

Generating Bundle File with Invocation Image deislabs/porter-kubernetes@sha256:d8aa654f5e60d64f698d79664480500b8de469a22e15dc69806e8172848e17d6 =======>
Generating parameter definition porter-debug ====>
Generating credential kubeconfig ====>
Starting to copy image deislabs/porter-kubernetes@sha256:d8aa654f5e60d64f698d79664480500b8de469a22e15dc69806e8172848e17d6...
Completed image deislabs/porter-kubernetes@sha256:d8aa654f5e60d64f698d79664480500b8de469a22e15dc69806e8172848e17d6 copy
WARN[0005] reference for unknown type: application/vnd.cnab.config.v1+json
Bundle tag deislabs/porter-kubernetes-bundle:1.0 pushed successfully, with digest "sha256:57c34a53e84607562e396280563186759139454d1704c727180aac1819b75a4f"
```

Note: you can safely ignore the `WARN[0005] reference for unknown type: application/vnd.cnab.config.v1+json` message.

When this command is complete, your CNAB bundle manifest and invocation image will have been successfully pushed to the specified OCI registry. It can then be installed with the `porter install` command:

```
$ porter install --tag deislabs/porter-kube-bundle:1.0 -c kool-kred
installing kube-example...
executing porter install configuration from /cnab/app/porter.yaml
Install Hello World App
```

The bundle can also be pulled with specified digest:

```
$ porter install --tag deislabs/porter-kube-bundle@sha256:57c34a53e84607562e396280563186759139454d1704c727180aac1819b75a4f -c kool-kred
installing kube-example...
executing porter install configuration from /cnab/app/porter.yaml
Install Hello World App
```

The later example ensures immutability for your bundle. After you've initially run `porter publish`, your tagged reference, such as `deislabs/porter-kube-bundle:1.0` can be updated with subsequent `porter publish` commands. However, the digested version `deislabs/porter-kube-bundle@sha256:57c34a53e84607562e396280563186759139454d1704c727180aac1819b75a4f` will not change. If you'd like to publish different version of the bundle, you will need to update both the `invocationImage` and `tag` attributes and run `porter build` before running `porter publish` again.