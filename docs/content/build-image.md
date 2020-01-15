---
title: Building Invocation Images
descriptions: How does Porter build an Invocation Image?
---

When you build a Cloud Native Application Bundle (CNAB) with Porter, a bundle.json and an invocation image are created for you. How does Porter turn your _porter.yaml_ into an invocation image? This walkthrough will explain how Porter constructs the invocation image, including how mixins and other bundles allow you to compose functionality.

## Starting From Scratch

When you create a new bundle with Porter, your project is bootstrapped with a sample _porter.yaml_ and a new _cnab_ directory. This scaffolding provides almost everything you need to generate your CNAB, including the invocation image. Let's use this to explain how the invocation image is built. 

To create a new CNAB with Porter, you first run `porter create`. The generated `porter.yaml` will look like this:

```yaml
# This is the configuration for Porter
# You must define steps for each action, but the rest is optional
# See https://porter.sh/author-bundles for documentation on how to configure your bundle
# Uncomment out the sections below to take full advantage of what Porter can do!

name: HELLO
version: 0.1.0
description: "An example Porter configuration"
# TODO: update the registry to your own, e.g. myregistry/porter-hello:v0.1.0
tag: getporter/porter-hello:v0.1.0

# Uncomment the line below to set a specific name for the invocation image
#invocationImage: getporter/porter-hello-installer:0.1.0

# Uncomment the line below to use a template Dockerfile for your invocation image
#dockerfile: Dockerfile.tmpl

mixins:
  - exec

install:
  - exec:
      description: "Install Hello World"
      command: bash
      flags:
        c: echo Hello World

upgrade:
  - exec:
      description: "World 2.0"
      command: bash
      flags:
        c: echo World 2.0

uninstall:
  - exec:
      description: "Uninstall Hello World"
      command: bash
      flags:
        c: echo Goodbye World


# See https://porter.sh/author-bundles/#dependencies
#dependencies:
#  mysql:
#    tag: getporter/mysql:v0.1.1
#    parameters:
#      database-name: wordpress

# See https://porter.sh/wiring/#credentials
#credentials:
#  - name: kubeconfig
#    path: /root/.kube/config

```

After the scaffolding is created, you may edit the _porter.yaml_ and modify the `tag: getporter/porter-hello:v0.1.0` element representing the bundle tag to include a Docker registry that you can push to. You may also uncomment and modify the `invocationImage: getporter/porter-hello:0.1.0` element representing the invocation image name to your liking. Note that the invocation image is not pushed during the `porter build` workflow.

Once you have modified the `porter.yaml`, you can run `porter build` to generate your first invocation image.  Here we add the `--verbose` flag to see all of the output:

```console
$ porter build --verbose
Copying porter runtime ===>
Copying mixins ===>
Copying mixin exec ===>

Generating Dockerfile =======>
FROM debian:stretch

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates

# exec mixin has no buildtime dependencies


COPY . $BUNDLE_DIR
RUN rm -fr $BUNDLE_DIR/.cnab
COPY .cnab /cnab
COPY porter.yaml $BUNDLE_DIR/porter.yaml
WORKDIR $BUNDLE_DIR
CMD ["/cnab/app/run"]

Writing Dockerfile =======>
FROM debian:stretch

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates

# exec mixin has no buildtime dependencies


COPY . $BUNDLE_DIR
RUN rm -fr $BUNDLE_DIR/.cnab
COPY .cnab /cnab
COPY porter.yaml $BUNDLE_DIR/porter.yaml
WORKDIR $BUNDLE_DIR
CMD ["/cnab/app/run"]

Starting Invocation Image Build =======>
Step 1/9 : FROM debian:stretch
 ---> 5c43e435cc11
Step 2/9 : ARG BUNDLE_DIR
 ---> Using cache
 ---> 7b7947fb2576
Step 3/9 : RUN apt-get update && apt-get install -y ca-certificates
 ---> Using cache
 ---> d60d94e3f701
Step 4/9 : COPY . $BUNDLE_DIR
 ---> 5493aa2241d3
Step 5/9 : RUN rm -fr $BUNDLE_DIR/.cnab
 ---> Running in f8bc113e739a
 ---> 88ea643205d0
Step 6/9 : COPY .cnab /cnab
 ---> 9c6c895f590b
Step 7/9 : COPY porter.yaml $BUNDLE_DIR/porter.yaml
 ---> 6f79f7b13e79
Step 8/9 : WORKDIR $BUNDLE_DIR
 ---> Running in 15799ffc05e8
 ---> c40ff2f77f45
Step 9/9 : CMD ["/cnab/app/run"]
 ---> Running in 76ff0004ec8e
 ---> e304c0fc1a25
Successfully built e304c0fc1a25
Successfully tagged jeremyrickard/porter-hello:0.1.0
```

A lot just happened by running that command! Let's walk through the output and discuss what happened.

```console
Copying porter runtime ===>
Copying mixins ===>
Copying mixin exec ===>
```

The first thing that happens after running `porter build`, Porter will copy its runtime plus any mixins into the `.cnab/app` directory of your bundle. 

Porter locates available mixins in the `$PORTER_HOME/mixins` directory. By default, the Porter home directory is located in `~/.porter`. In this example, we are using the `exec` mixin, so the `$PORTER_HOME/mixins/exec` directory will be copied into the invocation image. When a mixin is [installed](#tbd) for use with Porter, it contains binaries for multiple operating systems. The correct binary will be copied into the current `.cnab` directory for use in the invocation image.

After copying any mixins to the `.cnab` directory of the bundle, a Dockerfile is generated:

```console
Generating Dockerfile =======>
FROM debian:stretch

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates

# exec mixin has no buildtime dependencies


COPY . $BUNDLE_DIR
RUN rm -fr $BUNDLE_DIR/.cnab
COPY .cnab /cnab
COPY porter.yaml $BUNDLE_DIR/porter.yaml
WORKDIR $BUNDLE_DIR
CMD ["/cnab/app/run"]
```

Porter starts the Dockerfile by using a base image. You can customize the base image by specifying a Dockerfile template in the **porter.yaml**. Next, a set of CA certificates is added.  Next, contents of the current directory are copied into `/cnab/app/` in the invocation image. This will include any contributions from the mixin executables. Finally, an entry point that conforms to the CNAB specification is added to the image.

Once this is completed, the image is built:

```console
Starting Invocation Image Build =======>
Step 1/9 : FROM debian:stretch
 ---> 5c43e435cc11
Step 2/9 : ARG BUNDLE_DIR
 ---> Using cache
 ---> 7b7947fb2576
Step 3/9 : RUN apt-get update && apt-get install -y ca-certificates
 ---> Using cache
 ---> d60d94e3f701
Step 4/9 : COPY . $BUNDLE_DIR
 ---> 79290bcf128f
Step 5/9 : RUN rm -fr $BUNDLE_DIR/.cnab
 ---> Running in 7f12cd3f447d
 ---> 01b633a31bf8
Step 6/9 : COPY .cnab /cnab
 ---> 25c0b1e5f70a
Step 7/9 : COPY porter.yaml $BUNDLE_DIR/porter.yaml
 ---> dbb26cacf8d8
Step 8/9 : WORKDIR $BUNDLE_DIR
 ---> Running in b051cb2b6ddb
 ---> e10d6ab60595
Step 9/9 : CMD ["/cnab/app/run"]
 ---> Running in 50f1aa7c5b53
 ---> c8e0fc788a0d
Successfully built c8e0fc788a0d
Successfully tagged jeremyrickard/porter-hello-installer:0.1.0
```

## Mixins Help The Build

In the simple example above, the resulting Dockerfile was built entirely by the default `porter build` functionality. The `porter build` output reported that the `exec` mixin did not have any build time dependencies:

```
# exec mixin has no buildtime dependencies
```

In many cases, however, mixins will have build time requirements. Next let's see what happens when we use the Helm mixin. Here is another example `porter.yaml`:

```yaml
mixins:
- helm

name: mysql
version: "0.1.0"
tag: jeremyrickard/mysql:v0.1.0

credentials:
- name: kubeconfig
  path: /root/.kube/config

install:
- helm:
    description: "Install MySQL"
    name: porter-ci-mysql
    chart: stable/mysql
    version: "0.10.2"
uninstall:
- helm:
    description: "Uninstall MySQL"
    releases:
    - porter-ci-mysql
    purge: true
```

When we run `porter build` on this, the output is different:

```console
$ porter build --verbose
Copying porter runtime ===>
Copying mixins ===>
Copying mixin helm ===>

Generating Dockerfile =======>
FROM debian:stretch

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates

RUN apt-get update && \
 apt-get install -y curl && \
 curl -o helm.tgz https://get.helm.sh/helm-v2.14.3-linux-amd64.tar.gz && \
 tar -xzf helm.tgz && \
 mv linux-amd64/helm /usr/local/bin && \
 rm helm.tgz
RUN helm init --client-only
RUN apt-get update && \
 apt-get install -y apt-transport-https curl && \
 curl -o kubectl https://storage.googleapis.com/kubernetes-release/release/v1.15.3/bin/linux/amd64/kubectl && \
 mv kubectl /usr/local/bin && \
 chmod a+x /usr/local/bin/kubectl

COPY . $BUNDLE_DIR
RUN rm -fr $BUNDLE_DIR/.cnab
COPY .cnab /cnab
COPY porter.yaml $BUNDLE_DIR/porter.yaml
WORKDIR $BUNDLE_DIR
CMD ["/cnab/app/run"]
```

First, the `helm` mixin is copied instead of `exec` mixin. The Dockerfile looks similar in the beginning, but we can then see our next difference. The following lines of our generated Dockerfile were contributed by the `helm` mixin:

```
RUN apt-get update && \
 apt-get install -y curl && \
 curl -o helm.tgz https://get.helm.sh/helm-v2.14.3-linux-amd64.tar.gz && \
 tar -xzf helm.tgz && \
 mv linux-amd64/helm /usr/local/bin && \
 rm helm.tgz
RUN helm init --client-only
RUN apt-get update && \
 apt-get install -y apt-transport-https curl && \
 curl -o kubectl https://storage.googleapis.com/kubernetes-release/release/v1.15.3/bin/linux/amd64/kubectl && \
 mv kubectl /usr/local/bin && \
 chmod a+x /usr/local/bin/kubectl
```

How did that happen? To find out, let's first look at the `helm` mixin:

```console
~/.porter/mixins/helm/helm
A helm mixin for porter 👩🏽‍✈️

Usage:
  helm [command]

Available Commands:
  build       Generate Dockerfile lines for the bundle invocation image
  help        Help about any command
  install     Execute the install functionality of this mixin
  invoke      Execute the invoke functionality of this mixin
  schema      Print the json schema for the mixin
  uninstall   Execute the uninstall functionality of this mixin
  upgrade     Execute the upgrade functionality of this mixin
  version     Print the mixin version

Flags:
      --debug   Enable debug logging
  -h, --help    help for helm

Use "helm [command] --help" for more information about a command.
```

The [Porter Mixin Contract](#tbd) specifies that mixins must provide a `build` sub command that generates Dockerfile lines to support the runtime execution of the mixin. In the case of the `helm` mixin, this includes installing Helm and running a `helm init --client-only` to prepare the image. At build time, Porter uses the _porter.yaml_ to determine what mixins are required for the bundle. Porter then invokes the build sub-command for each specified mixin and appends that output to the base Dockerfile.

In the end, the result is a single invocation image with all of the necessary pieces: the porter-runtime, selected mixins and any relevant configuration files, scripts, charts or manifests. That invocation image can then be executed by any tool that supports the CNAB spec, while still taking advantage of the Porter capabilities.