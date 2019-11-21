---
title: Porter Client Docker Image
description: How to use the getporter/porter docker image
---

The [getporter/porter][porter] docker image provides the porter client installed in a
container.

It has tags that match what is available from our [install](/install/) page:
`latest`, `canary` and specific versions such as `v0.20.0-beta.1`.

**Notes**

* The docker socket must be mounted to the container in order to execute a
  bundle, using `-v /var/run/docker.sock:/var/run/docker.sock`.
* The `ENTRYPOINT` is set to `porter`, to change that you can use 
  `--entrypoint`, for example `docker run --rm -it --entrypoint /bin/sh porter`. 
* Don't mount the entire porter home directory, because that's where the porter
  binary is located, instead mount individual directories such as claims or
  credentials underneath it. Otherwise you will get an error like 
  `exec user process caused "exec format error"`.

## Examples
Here are some examples of how to use the porter client docker image.

### Create
```
docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v `pwd`/hello:/tmp/hello \
    -w /tmp/hello \
    getporter/porter create
```

Breaking down the command, here's what it just did:

* Mount the docker socket
* Mount a location from our local machine so that we can persist our bundle's files
* Set the working directory to the bundle directory
* Run the `porter create` command

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
    getporter/porter install -t getporter/hello:0.1.0

installing hello...
executing install action from hello (bundle instance: hello)
Install Hello World
Hello World
execution completed successfully!
```

### List
We can also ist our installed bundles with their status:

```
$ docker run -it --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $HOME/.porter/claims:/root/.porter/claims \
    getporter/porter list

NAME      CREATED         MODIFIED        LAST ACTION   LAST STATUS
hello     2 minutes ago   2 minutes ago   install       success
```

[porter]: https://hub.docker.com/r/getporter/porter/tags