---
title: "Using Docker inside Porter Bundles"
description: "Now you can use Docker and Docker Compose from inside Porter bundles"
date: "2020-04-23"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://carolynvanslyck.com/"
authorimage: "https://github.com/carolynvs.png"
image: "images/porter-with-docker-twitter-card.png"
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
* [Use Docker](#use-docker)

Here's the [full working example whalesay bundle][whalesay-bundle] for you to
follow along with.

[whalesay-bundle]: /examples/src/docker/

### Require Docker

The user running the bundle, and Porter, needs to know that this bundle
requires the local Docker daemon connected to the bundle. You need to a new
section to porter.yaml for required extensions, and defined a new prototype
extension that says that the bundle [requires access to a Docker
daemon](/author-bundles/#docker):

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

You can include the Docker CLI in your bundle by using the [docker mixin].

```yaml
mixins:
  - docker
```

You can then use the respective CLIs through the mixin:

```yaml
install:
- docker:
    description: "Run my container"
    run:
      image: hello-world
      rm: true
```

This blog post focuses on just the docker mixin, but here is a [full
working example for how to use Docker Compose in a
bundle](/examples/src/compose/).

### Use Docker

In my porter.yaml, I can use the docker mixin to execute docker commands
in my bundle:

**porter.yaml**

```yaml
name: examples/whalesay
version: 0.2.0
description: "An example bundle that uses docker through the magic of whalespeak"
registry: ghcr.io/getporter

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
  - docker

install:
  - docker:
      run:
        image: "docker/whalesay:latest"
        rm: true
        arguments:
          - cowsay
          - Hello World

upgrade:
  - docker:
      run:
        image: "docker/whalesay:latest"
        rm: true
        arguments:
          - cowsay
          - World 2.0

say:
  - docker:
      run:
        image: "docker/whalesay:latest"
        rm: true
        arguments:
          - cowsay
          - - ${ bundle.parameters.msg }

uninstall:
  - docker:
      run:
        image: "docker/whalesay:latest"
        rm: true
        arguments:
          - cowsay
          - Goodbye World
```

After I test the bundle and verify that it's ready for release, I use `porter publish` to push the new image `ghcr.io/getporter/examples/whalesay:v0.2.0` to the registry.

## Run that bundle

Now that the bundle is ready to use, the user running the bundle needs to
give the bundle elevated permission with the [Allow Docker Host
Access](/configuration/#allow-docker-host-access) setting. This
is because giving a container access to the host's Docker socket, or running a
container with `--privileged`, has security implications for the underlying host,
and should only be given to trusted containers, or in this case trusted bundles.

Let the whales speak!

```console
$ porter install --reference ghcr.io/getporter/examples/whalesay:v0.2.0 --allow-docker-host-access
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

Now let's see what else you can do with whalesay:

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
us to take this further, please reach out on the [porter][porter-repo] or [docker mixin][docker-repo]
repositories!

[porter-repo]: https://github.com/getporter/porter/
[docker-repo]: https://github.com/getporter/mixin-docker/
[compose-spec]: https://www.compose-spec.io/
[docker mixin]: /mixins/docker/
