---
title: "Using Docker inside Porter Bundles"
description: "Now you can use Docker and Docker Compose from inside Porter bundles"
date: "2020-04-23"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://twitter.com/carolynvs"
image: "images/porter-with-docker.png"
tags: ["docker", "mixins"]
---

<img src="/images/porter-with-docker.png" width="250px" align="right"/>

Sometimes you need a hammer, and that hammer happens to be a whale 🐳. We all
use containers as part of our pipeline: building images, running a one-off
command in a utility container, spinning up a test environment to verify your
application, or even more creative tasks that you have already
containerized. Well now you can reuse all that hard work and logic from within
your bundles!

Let's walk through using my favorite container, [docker/whalesay][whalesay], in a bundle. 

```
 _____________________
< Challenge Accepted! >
 ---------------------
    \
     \
      \
                    ##        .
              ## ## ##       ==
           ## ## ## ##      ===
       /""""""""""""""""___/ ===
  ~~~ {~~ ~~~~ ~~~ ~~~~ ~~ ~ /  ===- ~~~
       \______ o          __/
        \    \        __/
          \____\______/
```

[whalesay]: https://hub.docker.com/r/docker/whalesay/

## Author the bundle
Writing a bundle that uses Docker has a few steps:

* [Require Docker](#require-docker)
* [Install Docker](#install-docker)
* [Alternative - The docker-compose mixin](#the-docker-compose-mixin)
* [Use Docker](#use-docker)

Here's the [full working example whalesay bundle][whalesay-bundle] for you to
follow along with.

[whalesay-bundle]: https://github.com/deislabs/porter/tree/master/examples/docker

### Require Docker

The user running the bundle, and Porter, needs to know that this bundle
requires the local Docker daemon connected to the bundle. We have added a new
section to porter.yaml for required extensions, and defined a new prototype
extension that says that the bundle [requires access to a Docker
daemon](https://porter.sh/author-bundles/#docker):

```yaml
required:
- docker
```

Optionally you can include that you also want the bundle to be run with
`--privileged` which is intended for docker-in-docker scenarios:

```yaml
required:
- docker:
    privileged: true
```

### Install Docker

We can install the Docker CLI, or whatever is needed, into the
bundle using a [custom Dockerfile](https://porter.sh/custom-dockerfile/).

In our [Dockerfile.tmpl](https://porter.sh/src/examples/docker/Dockerfile.tmpl)
below, I am installing the Docker CLI and copying my files into the bundle:

```dockerfile
FROM debian:stretch

ARG BUNDLE_DIR
RUN apt-get update && apt-get install -y curl ca-certificates

ARG DOCKER_VERSION=19.03.8
RUN curl -o docker.tgz https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz && \
    tar -xvf docker.tgz && \
    mv docker/docker /usr/bin/docker && \
    chmod +x /usr/bin/docker && \
    rm docker.tgz

# Use the BUNDLE_DIR build argument to copy files into the bundle
COPY . $BUNDLE_DIR
```

### The docker-compose mixin

Alternatively, we have also created a new mixin,
[docker-compose](https://porter.sh/mixins/docker-compose/), that helps you
quickly use Docker Compose from inside your bundle. It handles installing the
Docker Compose CLI and gives you the following syntax:

```yaml
docker-compose:
  description: "Start Test Services"
  arguments:
  - up
  - -d
```

We are going to focus on just using Docker for this blog post, but here is a [full
working example for how to use Docker Compose in a
bundle](https://github.com/deislabs/porter-docker-compose/tree/master/examples/compose).

### Use Docker

Since we don't have a docker mixin, I am using the exec mixin to call the Docker CLI.
For ease of testing, I put my commands into a helper script, `helpers.sh`. This lets
me test out my commands locally without running the entire bundle.

**helpers.sh**

```bash
#!/usr/bin/env bash
set -euo pipefail

whalesay() {
  docker run --rm docker/whalesay:latest cowsay $1
}

install() {
  whalesay "Hello World"
}

upgrade() {
  whalesay "World 2.0"
}

uninstall() {
  whalesay "Goodbye World"
}

# Call the requested function and pass the arguments as-is
"$@"
```

Now in my porter.yaml I have very straightforward calls to the functions
defined in my helpers.sh script for each action:

**porter.yaml**

```yaml
name: whalesay
version: 0.1.1
description: "An example bundle that uses docker through the magic of whalespeak"
tag: getporter/whalesay:v0.1.1

dockerfile: Dockerfile.tmpl

required:
- docker

parameters:
- name: msg
  description: a message for the whales to speak
  type: string
  default: "whale hello there!"
  applyTo:
  - say

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

say:
- exec:
    description: "Say Something"
    command: ./helpers.sh
    arguments:
    - whalesay
    - '"{{ bundle.parameters.msg }}"'

uninstall:
- exec:
    description: "Uninstall Hello World"
    command: ./helpers.sh
    arguments:
    - uninstall
```

After I have tested the bundle, I used `porter publish` to push it up to `getporter/whalesay:v0.1.1`.

## Run that bundle

Now that the bundle is ready to use, the user running the bundle needs to
give the bundle elevated permission with the [Allow Docker Host
Access](https://porter.sh/configuration/#allow-docker-host-access) setting. This
is because giving a container access to the host's Docker socket, or running a
container with `--privileged`, has security implications for the underlying host,
and should only be given to trusted containers, or in this case trusted bundles.

Let the whales speak!

```console
$ porter install --tag getporter/whalesay:v0.1.1 --allow-docker-host-access
installing whalesay...
executing install action from whalesay (bundle instance: whalesay)
Install Hello World
 _____________
< Hello World >
 -------------
    \
     \
      \
                    ##        .
              ## ## ##       ==
           ## ## ## ##      ===
       /""""""""""""""""___/ ===
  ~~~ {~~ ~~~~ ~~~ ~~~~ ~~ ~ /  ===- ~~~
       \______ o          __/
        \    \        __/
          \____\______/
execution completed successfully!
```

I can set the flag `--allow-docker-host-access` with the `PORTER_ALLOW_DOCKER_HOST_ACCESS` environment variable so that I don't have to specify it for every command.

```console
export PORTER_ALLOW_DOCKER_HOST_ACCESS=true
```

Now let's see what else we can do with whalesay:

```console
$ porter invoke whalesay --action=say --param 'msg=try it yourself!'
invoking custom action say on whalesay...
executing say action from whalesay (bundle instance: whalesay)
Say Something
 __________________
< try it yourself! >
 ------------------
    \
     \
      \
                    ##        .
              ## ## ##       ==
           ## ## ## ##      ===
       /""""""""""""""""___/ ===
  ~~~ {~~ ~~~~ ~~~ ~~~~ ~~ ~ /  ===- ~~~
       \______ o          __/
        \    \        __/
          \____\______/
execution completed successfully!
```

This is hopefully just the first step towards first class support for Docker and
Docker Compose within Porter bundles, especially now that [Docker Compose has an
open specification][compose-spec]. If you are interested in collaborating with
us to take this further, please reach out on the [porter][porter-repo] or
[porter-docker-compose][compose-repo] repositories!

[porter-repo]: https://github.com/deislabs/porter/
[compose-repo]: https://github.com/deislabs/porter-docker-compose/
[compose-spec]: https://www.compose-spec.io/
