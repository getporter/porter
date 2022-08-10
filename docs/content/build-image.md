---
title: Building Invocation Images
descriptions: How does Porter build an Invocation Image?
---

When you build a Cloud Native Application Bundle (CNAB) with Porter, a bundle.json and an invocation image are created for you. How does Porter turn your _porter.yaml_ into an invocation image? This walkthrough will explain how Porter constructs the invocation image, including how mixins and other bundles allow you to compose functionality.

## Starting From Scratch

When you create a new bundle with Porter, your project is bootstrapped with a sample porter.yaml. This scaffolding provides almost everything you need to generate your CNAB, including the invocation image. Let's use this to explain how the invocation image is built. 

To create a new CNAB with Porter, you first run `porter create`. The generated porter.yaml will look like this:

```yaml
name: porter-hello
version: 0.1.0
description: "An example Porter configuration"
registry: getporter

mixins:
  - exec

install:
  - exec:
      description: "Install Hello World"
      command: ./helpers.sh
      arguments:
        - install

upgrade:
  - exec:
      description: "World 2.0"
      command: ./helpers.sh
      arguments:
        - upgrade

uninstall:
  - exec:
      description: "Uninstall Hello World"
      command: ./helpers.sh
      arguments:
        - uninstall
```

After the scaffolding is created, you may edit the porter.yaml and modify the `registry: localhost:5000` element representing the Docker registry that you can push to. Note that the bundle is not pushed during porter build.

Once you have modified the porter.yaml, you can run `porter build --debug` to generate your first invocation image.

```console
$  porter build --debug
Resolved porter binary from /usr/local/bin/porter to /Users/sigje/.porter/porter
Running linters for each mixin used in the manifest...
Copying porter runtime ===>
Copying mixins ===>
Copying mixin exec ===>

Generating Dockerfile =======>
DEBUG name:    exec
DEBUG pkgDir: /Users/sigje/.porter/mixins/exec
DEBUG file:
DEBUG stdin:
actions:
  install:
  - exec:
      arguments:
      - install
      command: ./helpers.sh
      description: Install Hello World2
  uninstall:
  - exec:
      arguments:
      - uninstall
      command: ./helpers.sh
      description: Uninstall Hello World
  upgrade:
  - exec:
      arguments:
      - upgrade
      command: ./helpers.sh
      description: World 2.0

/Users/sigje/.porter/mixins/exec/exec build --debug
FROM --platform=linux/amd64 debian:stretch-slim

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates

# exec mixin has no buildtime dependencies


COPY . ${BUNDLE_DIR}
RUN rm -fr ${BUNDLE_DIR}/.cnab
COPY .cnab /cnab
COPY porter.yaml ${BUNDLE_DIR}/porter.yaml
WORKDIR ${BUNDLE_DIR}
CMD ["/cnab/app/run"]

Writing Dockerfile =======>
FROM --platform=linux/amd64 debian:stretch-slim

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates

# exec mixin has no buildtime dependencies


COPY . ${BUNDLE_DIR}
RUN rm -fr ${BUNDLE_DIR}/.cnab
COPY .cnab /cnab
COPY porter.yaml ${BUNDLE_DIR}/porter.yaml
WORKDIR ${BUNDLE_DIR}
CMD ["/cnab/app/run"]

Starting Invocation Image Build =======>
Step 1/9 : FROM --platform=linux/amd64 debian:stretch-slim
 ---> 5738956efb6b
Step 2/9 : ARG BUNDLE_DIR
 ---> Using cache
 ---> c9d91881dd7c
Step 3/9 : RUN apt-get update && apt-get install -y ca-certificates
 ---> Using cache
 ---> afa85b98ed97
Step 4/9 : COPY . ${BUNDLE_DIR}
 ---> Using cache
 ---> e4057b41978c
Step 5/9 : RUN rm -fr ${BUNDLE_DIR}/.cnab
 ---> Using cache
 ---> ee114d95bc2d
Step 6/9 : COPY .cnab /cnab
 ---> Using cache
 ---> 1bb73c63ef65
Step 7/9 : COPY porter.yaml ${BUNDLE_DIR}/porter.yaml
 ---> Using cache
 ---> 483c6b05a0b7
Step 8/9 : WORKDIR ${BUNDLE_DIR}
 ---> Using cache
 ---> 9d2497296f3b
Step 9/9 : CMD ["/cnab/app/run"]
 ---> Using cache
 ---> 23c208fd5dc7
Successfully built 23c208fd5dc7
Successfully tagged getporter/porter-hello-installer:0.1.0
DEBUG name:    arm
DEBUG pkgDir: /Users/sigje/.porter/mixins/arm
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/arm/arm version --output json --debug
DEBUG name:    aws
DEBUG pkgDir: /Users/sigje/.porter/mixins/aws
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/aws/aws version --output json --debug
DEBUG name:    az
DEBUG pkgDir: /Users/sigje/.porter/mixins/az
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/az/az version --output json --debug
DEBUG name:    exec
DEBUG pkgDir: /Users/sigje/.porter/mixins/exec
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/exec/exec version --output json --debug
DEBUG name:    gcloud
DEBUG pkgDir: /Users/sigje/.porter/mixins/gcloud
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/gcloud/gcloud version --output json --debug
DEBUG name:    helm
DEBUG pkgDir: /Users/sigje/.porter/mixins/helm
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/helm/helm version --output json --debug
DEBUG name:    kubernetes
DEBUG pkgDir: /Users/sigje/.porter/mixins/kubernetes
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/kubernetes/kubernetes version --output json --debug
DEBUG name:    terraform
DEBUG pkgDir: /Users/sigje/.porter/mixins/terraform
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/terraform/terraform version --output json --debug
```

A lot just happened by running that command! Let's walk through the output and discuss what happened.

```console
Copying porter runtime ===>
Copying mixins ===>
Copying mixin exec ===>
```

First, Porter copies its runtime plus any mixins into the `.cnab/app` directory of your bundle. 

Porter locates available mixins in the $PORTER_HOME/mixins directory. By default, the Porter home directory is located in ~/.porter. In this example, we are using the exec mixin, so the $PORTER_HOME/mixins/exec directory will be copied into the invocation image. When a mixin is installed, it contains binaries for multiple operating systems. The correct binary will be copied into the bundle's .cnab directory for use in the invocation image.

After copying any mixins to the .cnab directory, a Dockerfile is generated:

```console
Generating Dockerfile =======>
FROM --platform=linux/amd64 debian:stretch

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates

# exec mixin has no buildtime dependencies


COPY . ${BUNDLE_DIR}
RUN rm -fr ${BUNDLE_DIR}/.cnab
COPY .cnab /cnab
COPY porter.yaml ${BUNDLE_DIR}/porter.yaml
WORKDIR ${BUNDLE_DIR}
CMD ["/cnab/app/run"]
```

Porter starts the [Dockerfile](/bundle/custom-dockerfile) by using a base image. You can customize the base image by specifying a Dockerfile template in the porter.yaml. By default, Porter only targets a single os/architecture(linux/amd64) for invocation image. If you want to use other platform, feel free to change the platform flag in the generated Dockerfile template. Next, a set of CA certificates is added.  Next, contents of the current directory are copied into the bundle directory (/cnab/app) in the invocation image. This will include any contributions from the mixin executables. Finally, an entry point that conforms to the CNAB specification is added to the image.

Once this is completed, the image is built:

```console
Starting Invocation Image Build =======>
Step 1/9 : FROM --platform=linux/amd64 debian:stretch
 ---> 5c43e435cc11
Step 2/9 : ARG BUNDLE_DIR
 ---> Using cache
 ---> 7b7947fb2576
Step 3/9 : RUN apt-get update && apt-get install -y ca-certificates
 ---> Using cache
 ---> d60d94e3f701
Step 4/9 : COPY . ${BUNDLE_DIR}
 ---> 79290bcf128f
Step 5/9 : RUN rm -fr ${BUNDLE_DIR}/.cnab
 ---> Running in 7f12cd3f447d
 ---> 01b633a31bf8
Step 6/9 : COPY .cnab /cnab
 ---> 25c0b1e5f70a
Step 7/9 : COPY porter.yaml ${BUNDLE_DIR}/porter.yaml
 ---> dbb26cacf8d8
Step 8/9 : WORKDIR ${BUNDLE_DIR}
 ---> Running in b051cb2b6ddb
 ---> e10d6ab60595
Step 9/9 : CMD ["/cnab/app/run"]
 ---> Running in 50f1aa7c5b53
 ---> c8e0fc788a0d
Successfully built c8e0fc788a0d
Successfully tagged jeremyrickard/porter-hello-installer:0.1.0
```

## Mixins Help The Build

In the simple example above, the build output reported that the exec mixin did not have any build time dependencies:

```
# exec mixin has no buildtime dependencies
```

In many cases, however, mixins will have build time requirements. Next let's see what happens when we use the Helm mixin. Here is another example porter.yaml:

```yaml
mixins:
- helm3:
    repositories:
      bitnami:
        url: "https://charts.bitnami.com/bitnami"

name: mysql
version: "0.1.0"
registry: jeremyrickard

credentials:
- name: kubeconfig
  path: /home/nonroot/.kube/config

install:
- helm3:
    description: "Install MySQL"
    name: porter-ci-mysql
    chart: bitnami/mysql
    version: "6.14.2"
uninstall:
- helm3:
    description: "Uninstall MySQL"
    releases:
    - porter-ci-mysql
    purge: true
```

When we run porter build on this, the output is different:

```console
$ porter build --verbose
Copying porter runtime ===>
Copying mixins ===>
Copying mixin helm ===>

Generating Dockerfile =======>
FROM --platform=linux/amd64 debian:stretch-slim

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

COPY . ${BUNDLE_DIR}
RUN rm -fr ${BUNDLE_DIR}/.cnab
COPY .cnab /cnab
COPY porter.yaml ${BUNDLE_DIR}/porter.yaml
WORKDIR ${BUNDLE_DIR}
CMD ["/cnab/app/run"]
```

First, the helm mixin is copied instead of exec mixin. The Dockerfile looks similar in the beginning, but we can then see our next difference. The following lines of our generated Dockerfile were contributed by the helm mixin:

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

How did that happen? To find out, let's first look at the helm mixin:

```console
~/.porter/mixins/helm/helm
A helm mixin for porter

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

Porter mixins must provide a build command that generates Dockerfile lines to support the runtime execution of the mixin. In the case of the helm mixin, this includes installing Helm and running a `helm init --client-only` to prepare the image. At build time, Porter uses the porter.yaml to determine what mixins are required for the bundle. Then Porter invokes the build sub-command for each specified mixin and appends that output to the base Dockerfile.

In the end, the result is a single invocation image with the necessary pieces: the porter-runtime, selected mixins and any relevant configuration files, scripts, charts or manifests. That invocation image can then be executed by any tool that supports the CNAB spec, while still taking advantage of the Porter capabilities.