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
# Uncomment out the sections below to take full advantage of what Porter can do!

mixins:
  - exec

name: HELLO
version: "0.1.0"
invocationImage: porter-hello:latest

install:
  - exec:
      description: "Install Hello World"
      command: bash
      arguments:
        - -c
        - echo Hello World

uninstall:
  - exec:
      description: "Uninstall Hello World"
      command: bash
      arguments:
        - -c
        - echo Goodbye World

#dependencies:
#  - name: mysql
#    parameters:
#      database-name: wordpress

#credentials:
#  - name: kubeconfig
#    path: /root/.kube/config
```

After the scaffolding is created, edit the _porter.yaml_ and modify the `invocationImage: porter-hello:latest` element to include a Docker registry that you can push to.

Once you have modified the `porter.yaml`, you can run `porter build` to generate your first invocation image:

```console
$ porter build
Copying dependencies ===>
Copying mixins ===>
Copying mixin exec ===>
Copying mixin porter ===>

Generating Dockerfile =======>
[FROM quay.io/deis/lightweight-docker-go:v0.2.0 FROM debian:stretch COPY cnab/ /cnab/ COPY porter.yaml /cnab/app/porter.yaml CMD ["/cnab/app/run"] COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt # exec mixin has no buildtime dependencies ]

Writing Dockerfile =======>
FROM quay.io/deis/lightweight-docker-go:v0.2.0
FROM debian:stretch
COPY cnab/ /cnab/
COPY porter.yaml /cnab/app/porter.yaml
CMD ["/cnab/app/run"]
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# exec mixin has no buildtime dependencies


Starting Invocation Image Build =======>
Step 1/6 : FROM quay.io/deis/lightweight-docker-go:v0.2.0
 ---> acf6712d2918
Step 2/6 : FROM debian:stretch
 ---> de8b49d4b0b3
Step 3/6 : COPY cnab/ /cnab/
 ---> Using cache
 ---> 209f021564f0
Step 4/6 : COPY porter.yaml /cnab/app/porter.yaml
 ---> Using cache
 ---> 10740e93dd11
Step 5/6 : CMD ["/cnab/app/run"]
 ---> Using cache
 ---> 7da487955ba9
Step 6/6 : COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
 ---> Using cache
 ---> ee8ade0c612c
Successfully built ee8ade0c612c
Successfully tagged jeremyrickard/porter-hello:latest
The push refers to repository [docker.io/jeremyrickard/porter-hello]
937aeedb310f: Preparing
8f17ef8f8a30: Preparing
fc9479d629d7: Preparing
c581f4ede92d: Preparing
8f17ef8f8a30: Layer already exists
937aeedb310f: Layer already exists
fc9479d629d7: Pushed
c581f4ede92d: Pushed
latest: digest: sha256:c3187dc004475bd754235caf735d5adc449405126091594b24a38ebba93ae76a size: 1158

Generating Bundle File with Invocation Image jeremyrickard/porter-hello@sha256:c3187dc004475bd754235caf735d5adc449405126091594b24a38ebba93ae76a =======>
```

A lot just happened by running that command! Let's break walk through the output and discuss what happened.

```console
Copying dependencies ===>
Copying mixins ===>
Copying mixin exec ===>
Copying mixin porter ===>
```

The first thing that happens after running `porter build`, Porter will copy any dependencies and mixins into the `cnab\app` directory of your bundle. 

Porter locates available mixins in the `$PORTER_HOME\mixins` directory. By default, the Porter home directory is located in `~/.porter`. In this example, we are using the `exec` mixin, so the `$PORTER_HOME\mixins\exec` directory will be copied into the invocation image. When a mixin is [installed](#tbd) for use with Porter, it contains binaries for multiple operating systems. The correct binary will be copied into the current `cnab` directory for use in the invocation image.

After copying any dependencies and mixins to the `cnab` directory of the bundle, a Dockerfile is generated:

```console
Generating Dockerfile =======>
FROM quay.io/deis/lightweight-docker-go:v0.2.0
FROM debian:stretch
COPY cnab/ /cnab/
COPY porter.yaml/cnab/app/porter.yaml
CMD ["/cnab/app/run"]
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# exec mixin has no buildtime dependencies
```

Porter starts the Dockerfile by using a base image. The base image is currently not configurable. Next, the `cnab` directory is added to the image. This will include any contributions from dependencies and the mixin executables. Next, the _porter.yaml_ is added to the image. Next, an entry point that conforms to the CNAB specification is added to the image. Finally, a set of CA certificates is added.

Once this is completed, the image is built and pushed to the specified Docker registry:

```console
Starting Invocation Image Build =======>
Step 1/6 : FROM quay.io/deis/lightweight-docker-go:v0.2.0
 ---> acf6712d2918
Step 2/6 : FROM debian:stretch
 ---> de8b49d4b0b3
Step 3/6 : COPY cnab/ /cnab/
 ---> Using cache
 ---> 209f021564f0
Step 4/6 : COPY porter.yaml /cnab/app/porter.yaml
 ---> Using cache
 ---> 10740e93dd11
Step 5/6 : CMD ["/cnab/app/run"]
 ---> Using cache
 ---> 7da487955ba9
Step 6/6 : COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
 ---> Using cache
 ---> ee8ade0c612c
Successfully built ee8ade0c612c
Successfully tagged jeremyrickard/porter-hello:latest
The push refers to repository [docker.io/jeremyrickard/porter-hello]
937aeedb310f: Preparing
8f17ef8f8a30: Preparing
fc9479d629d7: Preparing
c581f4ede92d: Preparing
8f17ef8f8a30: Layer already exists
937aeedb310f: Layer already exists
fc9479d629d7: Pushed
c581f4ede92d: Pushed
latest: digest: sha256:c3187dc004475bd754235caf735d5adc449405126091594b24a38ebba93ae76a size: 1158
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
invocationImage: jeremyrickard/porter-mysql:latest

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
    name: porter-ci-mysql
    purge: true
```

When we run `porter build` on this, the output is different:

```console
$ porter build 
Copying dependencies ===>
Copying mixins ===>
Copying mixin helm ===>
Copying mixin porter ===>

Generating Dockerfile =======>
FROM quay.io/deis/lightweight-docker-go:v0.2.0 
FROM debian:stretch
COPY cnab/ /cnab/
COPY porter.yaml /cnab/app/porter.yaml
CMD ["/cnab/app/run"]
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
RUN apt-get update && \
 apt-get install -y curl && \
 curl -o helm.tgz https://storage.googleapis.com/kubernetes-helm/helm-v2.11.0-linux-amd64.tar.gz && \
 tar -xzf helm.tgz && \
 mv linux-amd64/helm /usr/local/bin && \
 rm helm.tgz
RUN helm init --client-only
```

First, the `helm` mixin is copied instead of `exec` mixin. The Dockerfile looks similar in the beginning, but we can then see our next difference. The following lines of our generated Dockerfile were contributed by the `helm` mixin:

```
RUN apt-get update && \
 apt-get install -y curl && \
 curl -o helm.tgz https://storage.googleapis.com/kubernetes-helm/helm-v2.11.0-linux-amd64.tar.gz && \
 tar -xzf helm.tgz && \
 mv linux-amd64/helm /usr/local/bin && \
 rm helm.tgz
RUN helm init --client-only
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
  install     Execute the install functionality of this mixin
  version     Print the mixin version

Flags:
      --debug   Enable debug logging
  -h, --help    help for helm

Use "helm [command] --help" for more information about a command.
```

The [Porter Mixin Contract](#tbd) specifies that mixins must provide a `build` sub command that generates Dockerfile lines to support the runtime execution of the mixin. In the case of the `helm` mixin, this includes installing Helm and running a `helm init --client-only` to prepare the image. At build time, Porter uses the _porter.yaml_ to determine what mixins are required for the bundle. Porter then invokes the build sub-command for each specified mixin and appends that output to the base Dockerfile.

## Including The Dependencies

Now that we've seen how Porter utilizes mixins to build an invocation image, it is time to address how dependencies are included. Consider the following sample `porter.yaml`:

```yaml
mixins:
- exec 

name: dependency-example 
version: "0.1.0"
invocationImage: jeremyrickard/dependency-example:latest

dependencies:
- name: mysql
  parameters:
    database_name: wordpress
    mysql_user: wordpress

install:
- exec:
    description: "Say Hello"
    command: bash
    arguments:
      - -c
      - echo Hello World

uninstall:
- exec:
    description: "Say Goodbye"
    command: bash
    arguments:
      - -c
      - echo Goodbye World
```

This bundle, for example, declares a dependency on a bundle named `mysql`. The CNAB specification doesn't provide a mechanism for handling dependency resolution. Porter supplements the CNAB spec to support dependencies by resolving any dependencies at build time, including the contents of each dependency in the invocation image. At runtime the contents of that bundle will therefore be in the bundle and the Porter runtime component can execute them successfully.

```console
$ porter build
Copying dependencies ===>
Copying bundle dependency mysql ===>
Copying mixins ===>
Copying mixin helm ===>
Copying mixin exec ===>
Copying mixin porter ===>

Generating Dockerfile =======>
[FROM quay.io/deis/lightweight-docker-go:v0.2.0 FROM debian:stretch COPY cnab/ /cnab/ COPY porter.yaml /cnab/app/porter.yaml CMD ["/cnab/app/run"] COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt RUN apt-get update && \  apt-get install -y curl && \  curl -o helm.tgz https://storage.googleapis.com/kubernetes-helm/helm-v2.11.0-linux-amd64.tar.gz && \  tar -xzf helm.tgz && \  mv linux-amd64/helm /usr/local/bin && \  rm helm.tgz RUN helm init --client-only # exec mixin has no buildtime dependencies ]

Writing Dockerfile =======>
FROM quay.io/deis/lightweight-docker-go:v0.2.0
FROM debian:stretch
COPY cnab/ /cnab/
COPY porter.yaml /cnab/app/porter.yaml
CMD ["/cnab/app/run"]
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
RUN apt-get update && \
 apt-get install -y curl && \
 curl -o helm.tgz https://storage.googleapis.com/kubernetes-helm/helm-v2.11.0-linux-amd64.tar.gz && \
 tar -xzf helm.tgz && \
 mv linux-amd64/helm /usr/local/bin && \
 rm helm.tgz
RUN helm init --client-only
# exec mixin has no buildtime dependencies
```

As we can see from this output, Porter found the `mysql` dependency and copied it into our bundle. If we examine the `cnab` directory after this build is complete, it should look like this:

```console
$ tree cnab/
cnab/
└── app
    ├── bundles
    │   └── mysql
    │       └── porter.yaml
    ├── mixins
    │   ├── exec
    │   │   ├── exec
    │   │   └── exec-runtime
    │   ├── helm
    │   │   ├── helm
    │   │   └── helm-runtime
    │   └── porter
    │       ├── porter
    │       └── porter-runtime
    ├── porter-runtime
    └── run
```

Porter found the `mysql` bundle by first looking in the `./bundles` directory. If nothing was found, it then checked `$PORTER_HOME\bundles` for the `mysql` bundle. In this case, it found the bundle and was able to copy it into the directory. Porter also copied the `helm` mixin into our bundle, despite the `porter.yaml` declaring only the `exec` mixin. It did this because the `mysql` dependency requires the `helm` mixin. As we can also see from the Dockerfile, it included the mixin for the `mysql` before the `exec` mixin.

In the end, regardless of whether there are dependencies, the result is a single invocation image with all of the necessary pieces: the porter-runtime, selected mixins and any relevant configuration files, scripts, charts or manifests. That invocation image can then be executed by any tool that supports the CNAB spec, while still taking advantage of the Porter capabilities.