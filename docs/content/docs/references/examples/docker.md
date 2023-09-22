---
title: "Example: Docker"
description: "Learn how to use Docker inside a bundle"
weight: 20
---

<img src="/images/porter-with-docker.png" width="250px" align="right"/>

Source: https://getporter.org/examples/src/docker

The [ghcr.io/getporter/examples/whalesay] bundle demonstrates how to use Docker inside a bundle!

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
requires the local Docker daemon connected to the bundle. We have added a new
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

### Use Docker

In my porter.yaml, I can use the docker mixin to execute docker commands
in my bundle:

<script src="https://gist-it.appspot.com/https://github.com/getporter/examples/blob/main/docker/porter.yaml"></script>

After I have tested the bundle, I used `porter publish` to push it up to `ghcr.io/getporter/examples/whalesay:v0.2.0`.

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

[docker mixin]: /mixins/docker/
[ghcr.io/getporter/examples/whalesay]: https://github.com/getporter/examples/tree/main/docker
