---
title: Porter Client Docker Image
description: How to use the ghcr.io/getporter/porter Docker image
---

The [ghcr.io/getporter/porter][porter] Docker image provides the Porter client installed in a
container. Mixins and plugins are **not** installed by default and must be mounted into /app/.porter.

It has tags that match what is available from our [install](/install/) page:
latest, canary and specific versions such as v0.20.0-beta.1.

**Notes**

* The Docker socket must be mounted to the container in order to execute a
  bundle, using `-v /var/run/docker.sock:/var/run/docker.sock`.
* The `ENTRYPOINT` is set to `porter`. To change this, you can use
  `--entrypoint`, e.g. `docker run --rm -it --entrypoint /bin/sh porter`.
* Don't mount the entire Porter home directory, because that's where the porter
  binary is located. Instead, mount individual directories such as mixins, claims,
  results, outputs, etc if needed. Otherwise, you will get an error
  like `exec user process caused "exec format error"`.

## Examples
Here are some examples of how to use the Porter client Docker image.

### Create
```
docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v `pwd`/hello:/tmp/hello \
    -w /tmp/hello \
    ghcr.io/getporter/porter create
```

Breaking down the command, here's what it just did:

* Mount the Docker socket.
* Mount a location from our local machine so that we can persist our bundle's files.
* Set the working directory to the bundle directory.
* Run the `porter create` command.

After this executes, you should be able to see the bundles files:

```
$ ls hello/
Dockerfile  Dockerfile.tmpl  README.md	porter.yaml
```

### Publish
Now let's publish the bundle to an OCI registry:

```
docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v `pwd`/hello:/tmp/hello \
    -w /tmp/hello \
    ghcr.io/getporter/porter publish
```

### Install
Finally, let's install a bundle:

```
$ docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $HOME/.porter/claims:/app/.porter/claims \
    -v $HOME/.porter/results:/app/.porter/results \
    -v $HOME/.porter/outputs:/app/.porter/outputs \
    ghcr.io/getporter/porter install -r ghcr.io/getporter/examples/porter-hello:0.2.0

installing hello...
executing install action from examples/porter-hello (installation: hello)
Install Hello World
Hello World
execution completed successfully!
```

### List
We can also list our installed bundles with their status:

```
$ docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $HOME/.porter/claims:/app/.porter/claims \
    -v $HOME/.porter/results:/app/.porter/results \
    -v $HOME/.porter/outputs:/app/.porter/outputs \
    getporter/porter list

NAME      CREATED         MODIFIED        LAST ACTION   LAST STATUS
hello     2 minutes ago   2 minutes ago   install       success
```

[porter]: https://github.com/getporter/porter/pkgs/container/porter
