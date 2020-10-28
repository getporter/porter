---
title: Custom Dockerfile
description: Defining a custom Dockerfile for your Porter bundle
---

Porter automatically generates a Dockerfile and uses it to build the invocation
image for your bundle. By default it copies all the files from the current
directory into the bundle, and installs SSL certificates. Sometimes you may want
to full control over your bundle's invocation image, for example to install
additional software used by the bundle.

When you run `porter create` template Dockerfile is created for you
in the current directory named **Dockerfile.tmpl**:

```Dockerfile
FROM debian:stretch

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates

# This is a template Dockerfile for the bundle's invocation image
# You can customize it to use different base images, install tools and copy configuration files.
#
# Porter will use it as a template and append lines to it for the mixins
# and to set the CMD appropriately for the CNAB specification.
#
# Add the following line to porter.yaml to instruct Porter to use this template
# dockerfile: Dockerfile.tmpl

# You can control where the mixin's Dockerfile lines are inserted into this file by moving "# PORTER_MIXINS" line
# another location in this file. If you remove that line, the mixins generated content is appended to this file.
# PORTER_MIXINS

# Use the BUNDLE_DIR build argument to copy files into the bundle
COPY . $BUNDLE_DIR

```

Add the following line to your **porter.yaml** file to instruct porter to use
the template, instead of generating one from scratch:

```yaml
dockerfile: Dockerfile.tmpl
```

It is your responsibility to provide a suitable base image, for example one that
has root ssl certificates installed. *You must use a base image that is
debian-based, such as `debian` or `ubuntu` with apt installed.* Mixins assume
that apt is available to install packages.

When using a Dockerfile template, you must manually copy any files you need in
your bundle using COPY statements. A few conventions are followed by Porter to
help with this task:

## BUNDLE_DIR

Your template must declare `ARG BUNDLE_DIR`, which is the path to the bundle
directory inside the invocation image. You may then use this when copying files
from the local filesystem:

```Dockerfile
COPY . $BUNDLE_DIR
```

## PORTER_MIXINS

The mixins used by your bundle generate Dockerfile lines that must be injected
into the Dockerfile template. You can control where they are injected by placing
a comment in your Dockerfile template:

```Dockerfile
# PORTER_MIXINS
```

When that line is omitted, the mixins Dockerfile lines are appended to the end
of the template.

The location of this comment can significantly impact the time it takes to
rebuild your bundle, due to image layers and caching. By default this line is
placed before copying your local files into the bundle, so that you can iterate
on your scripts and on the porter manifest without having to rebuild those
layers of the invocation image.
