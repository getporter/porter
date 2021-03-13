---
title: Porter Client Docker Image
description: How to use the getporter/porter Docker image
---

The [getporter/porter][porter] Docker image provides the Porter client installed in a
container.

It has tags that match what is available from our [install](/install/) page:
latest, canary and specific versions such as v0.20.0-beta.1.

**Notes**

* The Docker socket must be mounted to the container in order to execute a
  bundle, using `-v /var/run/docker.sock:/var/run/docker.sock`.
* The `ENTRYPOINT` is set to `porter`. To change this, you can use
  `--entrypoint`, e.g. `docker run --rm -it --entrypoint /bin/sh porter`.
* Don't mount the entire Porter home directory, because that's where the porter
  binary is located. Instead, mount individual directories such as claims,
  results and outputs (all three are used to record data for an installation)
  or credentials and parameters, if needed. Otherwise you will get an error
  like `exec user process caused "exec format error"`.

## Examples
Here are some examples of how to use the Porter client Docker image.

### Create
```
docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v `pwd`/hello:/tmp/hello \
    -w /tmp/hello \
    getporter/porter create
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
    getporter/porter publish
```

### Install
Finally let's install a bundle:

```
$ docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $HOME/.porter/claims:/root/.porter/claims \
    -v $HOME/.porter/results:/root/.porter/results \
    -v $HOME/.porter/outputs:/root/.porter/outputs \
    getporter/porter install -t getporter/porter-hello:0.1.0

installing hello...
executing install action from hello (installation: hello)
Install Hello World
Hello World
execution completed successfully!
```

### List
We can also list our installed bundles with their status:

```
$ docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $HOME/.porter/claims:/root/.porter/claims \
    -v $HOME/.porter/results:/root/.porter/results \
    -v $HOME/.porter/outputs:/root/.porter/outputs \
    getporter/porter list

NAME      CREATED         MODIFIED        LAST ACTION   LAST STATUS
hello     2 minutes ago   2 minutes ago   install       success
```

[porter]: https://hub.docker.com/r/getporter/porter/tags