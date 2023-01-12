---
title: Porter Workshop Docker Image
description: How to use the getporter/workshop Docker image
draft: true
---

The [ghcr.io/getporter/workshop][workshop] Docker image provides the Porter client installed in a
container that has Docker-in-Docker. It is suitable for workshops because
participants don't need to figure out their various Docker setups since they will
use the Docker host inside the container.

It has tags that match what is available from our [install](/install/) page:
latest, canary and specific versions such as v0.20.0-beta.1.

## Start the workshop container
```
docker run -d --privileged --name workshop ghcr.io/getporter/workshop
```

## Log into the workshop container
```
docker exec -it workshop sh
```

For the rest of the workshop, participants will execute commands from inside
the workshop container. 

**Notes**

You cannot mount volumes to arbitrary points into the workshop container from
your local computer because the image is based on the [docker:dind] image, see
[Where to Store Data](https://hub.docker.com/_/docker) for more information.

[workshop]: https://github.com/orgs/getporter/packages/container/package/workshop
