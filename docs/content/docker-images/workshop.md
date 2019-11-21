---
title: Porter Workshop Docker Image
description: How to use the getporter/workshop docker image
---

The [getporter/workshop][workshop] docker image provides the porter client installed in a
container that has Docker in Docker. It is suitable for workshops because
participants don't need to figure out their various docker setups since they will
use the docker host inside the container.

It has tags that match what is available from our [install](/install/) page:
`latest`, `canary` and specific versions such as `v0.20.0-beta.1`.

## Start the workshop container
```
docker run -d --privileged --name workshop getporter/workshop
```

## Log into the workshop container
```
docker exec -it workshop sh
```

For the rest of the workshop, participants will execute commands from inside
the workshop container. 

**Notes**

You cannot mount volumes into the workshop container from your local computer.

[workshop]: https://hub.docker.com/r/getporter/workshop/tags
