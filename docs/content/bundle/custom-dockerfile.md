---
title: Custom Dockerfile
description: Defining a custom Dockerfile for your Porter bundle
aliases:
- /custom-dockerfile/
---

Porter automatically generates a Dockerfile and uses it to build the invocation image for your bundle.
It runs the container as an unprivileged user with membership in the root group, copies all the files from the current directory into the bundle, and installs SSL certificates.
Sometimes you may want to full control over your bundle's invocation image, for example to install additional software used by the bundle.

When you run porter create, a template Dockerfile is created for you in the current directory named **template.Dockerfile**:

```Dockerfile
# syntax=docker/dockerfile-upstream:1.4.0
# This is a template Dockerfile for the bundle's invocation image
# You can customize it to use different base images, install tools and copy configuration files.
#
# Porter will use it as a template and append lines to it for the mixins
# and to set the CMD appropriately for the CNAB specification.
#
# Add the following line to porter.yaml to instruct Porter to use this template
# dockerfile: template.Dockerfile

# You can control where the mixin's Dockerfile lines are inserted into this file by moving the "# PORTER_*" tokens
# another location in this file. If you remove a token, its content is appended to the end of the Dockerfile.
FROM --platform=linux/amd64 debian:stretch-slim

# PORTER_INIT

RUN rm -f /etc/apt/apt.conf.d/docker-clean; echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache
RUN --mount=type=cache,target=/var/cache/apt --mount=type=cache,target=/var/lib/apt \
    apt-get update && apt-get install -y ca-certificates

# PORTER_MIXINS

# Use the BUNDLE_DIR build argument to copy files into the bundle's working directory
COPY --link . ${BUNDLE_DIR}
```

Add the following line to your **porter.yaml** file to instruct porter to use the template, instead of generating one from scratch:

```yaml
dockerfile: template.Dockerfile
```

It is your responsibility to provide a suitable base image, for example one that has root ssl certificates installed. 
*You must use a base image that is debian-based, such as debian or ubuntu with apt installed.*
Mixins assume that apt is available to install packages.
Porter only supports targeting a single os/architecture when the bundle is built. By default, Porter targets linux/amd64.
You can change the platform used in the Dockerfile.

# Buildkit

Porter automatically builds with Docker [buildkit] enabled.
The following docker flags are supported on the [porter build] command: \--ssh, \--secret, \--build-arg.
With these you can take advantage of Docker's support for using SSH connections, mounting secrets, and specifying custom build arguments.

By default, Porter uses the [1.4.0 dockerfile syntax](https://docs.docker.com/engine/reference/builder/#syntax), but you can modify this line to use new versions as they are released.

[buildkit]: https://docs.docker.com/develop/develop-images/build_enhancements/
[porter build]: /cli/porter_build/

# Bundles do not run as root

Porter runs the bundle image as a non-root user.
This means that if you need to initialize the user's home directory, you should use the [BUNDLE_USER](#BUNDLE_USER) build argument to locate the home directory.
If you need to run some commands as root and others as the non-root user that the bundle image will run under, you can use that same argument to switch the current user in the Dockerfile.
By default, all commands in the Dockerfile run as root (so that you can install and configure the image), so you will need to carefully use the USER statement to switch to the non-root user.

Here is a truncated and commented example of how the helm3 mixin performs some setup as root and some as the non-root user:
```Dockerfile
# Truncated, above we set up the image...

# The helm3 mixin performs system-level configuration, such as installing the helm cli
ENV HELM_EXPERIMENTAL_OCI=1
RUN apt-get update && apt-get install -y curl
RUN curl https://get.helm.sh/helm-v3.8.2-linux-amd64.tar.gz --output helm3.tar.gz
RUN tar -xvf helm3.tar.gz && rm helm3.tar.gz
RUN mv linux-amd64/helm /usr/local/bin/helm3
RUN curl -o kubectl https://storage.googleapis.com/kubernetes-release/release/v1.22.1/bin/linux/amd64/kubectl &&\
    mv kubectl /usr/local/bin && chmod a+x /usr/local/bin/kubectl

# The helm3 mixin switches temporarily to run a couple commands as the non-root user
USER ${BUNDLE_USER}
RUN helm3 repo add stable https://charts.helm.sh/stable
RUN helm3 repo update

# The helm3 mixin switches back to root so that the rest of the Dockerfile is run with a user with the correct permissions
USER root

# Truncated, continue setting up the bundle's image...
```

The non-root user matters both at build time when the image is built and also at runtime when it runs in a container.
The user that the container runs as is not configurable and any files written by the container are owned by the non-root user.
This is relevant when running a bundle in a Kubernetes pod (either with the kubernetes driver manually or with the Porter Operator).
It is important when mounting volumes into the pod that the non-root user has read/write access.

# Special Comments 
Porter uses comments as placeholders to inject lines into your Dockerfile that all Porter bundles require.
You can move the comment to another location in the file to optimize your Docker build times and layer caching.
If you omit the comment entirely, Porter will still inject the contents for that section into your Dockerfile, and we recommend keeping the comments in so that you can control where the contents are injected.

## PORTER_INIT

Porter includes additional Dockerfile lines that standardize all Porter bundles, such as declaring the BUNDLE_DIR argument, and creating a user for the bundle to run as.
You can control where these lines are injected by placing a comment in your Dockerfile template:

```Dockerfile
# PORTER_INIT
```

When that line is omitted, the lines are inserted after the FROM statement at the top of your template.

## PORTER_MIXINS

The mixins used by your bundle generate Dockerfile lines that must be injected into the Dockerfile template.
You can control where they are injected by placing a comment in your Dockerfile template:

```Dockerfile
# PORTER_MIXINS
```

When that line is omitted, the lines are appended to the end of the template.

The location of this comment can significantly impact the time it takes to rebuild your bundle, due to image layers and caching.
By default, this line is placed before copying your local files into the bundle, so that you can iterate on your scripts and on the porter manifest without having to rebuild those layers of the invocation image.


# Variables

When using a Dockerfile template, you must manually copy any files you need in your bundle using COPY statements.
A few conventions are followed by Porter to help with this task:

## BUNDLE_UID

The **BUNDLE_UID** argument declared in the [PORTER_INIT](#porter_init) section is the user id that the bundle's container runs as.
Below is an example of how to run a command as the bundle's user:

```Dockerfile
USER ${BUNDLE_UID}
RUN whoami
# 65532
```

## BUNDLE_GID

The **BUNDLE_GID** argument declared in the [PORTER_INIT](#porter_init) section is the group id that the bundle's container runs as.
Below is an example of how to copy a file into a directory outside [BUNDLE_DIR](#bundle_dir) and set the permissions so that the bundle can access them when it is run:

```Dockerfile
COPY --chown=${BUNDLE_UID}:${BUNDLE_GID} --chmod=770 myapp /myapp
```

## BUNDLE_USER

The **BUNDLE_USER** argument is the username that the bundle's container runs as.
Below is an example of how to copy files into the user's home directory:

```Dockerfile
COPY myfile /home/${BUNDLE_USER}/
```

## BUNDLE_DIR

The **BUNDLE_DIR** argument is declared in the [PORTER_INIT](#porter_init) section is the path to the bundle directory inside the invocation image.
You may then use this when copying files from the local filesystem to the bundle's working directory.
We strongly recommend that you always use this variable and do not copy files into directories outside BUNDLE_DIR.
If you do, you are responsible for setting the file permissions so that the bundle's user ([BUNDLE_USER](#bundle_user)) and the bundle's group (root) have the same permissions.

```Dockerfile
COPY . ${BUNDLE_DIR}
```

## See Also

* [Why you do not want to run containers as root](https://medium.com/@mccode/processes-in-containers-should-not-run-as-root-2feae3f0df3b)
* [Security Features](/security-features/)

[Buildkit]: https://docs.docker.com/develop/develop-images/build_enhancements/
[experimental]: /configuration/#experimental-feature-flags
[build-drivers]: /configuration/#build-drivers
