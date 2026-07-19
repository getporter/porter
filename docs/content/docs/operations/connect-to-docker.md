---
title: Connect to Docker
description: Configure Porter to authenticate and connect to a Docker engine
weight: 5
aliases:
  - /operations/connect-docker/
---

Some Porter commands connect to a Docker engine in order to build, push, and pull the bundle image.
Learn how to configure Porter to connect with various Docker configurations.

- [Connect to the local Docker engine](#connect-to-the-local-docker-engine)
- [Connect to a remote Docker engine](#connect-to-a-remote-docker-engine)
- [Additional Supported Docker Settings](#additional-supported-docker-settings)
- [Access the Docker host's network from a bundle](#access-the-docker-hosts-network-from-a-bundle)

## Connect to the local Docker engine

Porter defaults to connecting to the local Docker engine and no additional configuration is required.

Try it out by running `porter install --reference ghcr.io/getporter/examples/porter-hello:v0.2.0`.

## Connect to a remote Docker engine

Porter uses the standard Docker environment variables to connect to a remote Docker engine:

- **DOCKER_HOST**: The host name and port of the remote Docker engine.
- **DOCKER_CERT_PATH**: The local directory containing the certificates necessary to connect to the remote Docker engine. By default, Porter looks for ca.pem, cert.pem, key.pem in ~/.docker/certs.
- **DOCKER_TLS_VERIFY**: When this environment variable is set, Porter will use TLS to connect to the remote Docker engine. If the value is not true, the TLS certificates will not be verified.

Below is an example of how to set and use these environment variables.

**Bash**

```bash
export DOCKER_HOST="example.com:2376"
export DOCKER_TLS_VERIFY="true"
export DOCKER_CERT_PATH="/home/me/example-certs"

porter install --reference ghcr.io/getporter/examples/porter-hello:v0.2.0
```

**Powershell**

```powershell
$env:DOCKER_HOST="example.com:2376"
$env:DOCKER_TLS_VERIFY="true"
$env:DOCKER_CERT_PATH="C:\Users\me\example-certs"

porter install --reference ghcr.io/getporter/examples/porter-hello:v0.2.0
```

## Additional Supported Docker Settings

Porter supports additional Docker environment variables that may be useful to you:

- **DOCKER_NETWORK**: Specifies the name of an existing [Docker network] that Porter should use when running Docker containers. Set this to `host` to give the bundle access to the [Docker host's network](#access-the-docker-hosts-network-from-a-bundle).
- **DOCKER_CONTEXT**: Specifies the name of an existing [Docker context] that Porter should use when running Docker containers.
- **CLEANUP_CONTAINERS**: Controls whether Porter removes the Docker container used to run a bundle action once it finishes. Defaults to `true`. Set to `false` to leave the stopped container behind so you can inspect its filesystem, for example while authoring or debugging a bundle.

[Docker context]: https://docs.docker.com/engine/context/working-with-contexts/
[Docker network]: https://docs.docker.com/engine/reference/commandline/network/

### Inspect the container after a bundle action runs

By default Porter removes the container it used to run a bundle action once
the action finishes. Set `CLEANUP_CONTAINERS=false` to leave the stopped
container behind, then use Docker to look around:

```bash
CLEANUP_CONTAINERS=false porter install --reference ghcr.io/getporter/examples/porter-hello:v0.2.0
```

Find the container. Since it was just created, `docker ps -l` shows only the
most recently created container, which avoids having to scan through
unrelated containers on your machine:

```bash
docker ps -l
```

If something else was created after it, `docker ps -l` won't show the right
one. `docker ps -a` lists containers newest-first, so filter for the
`COMMAND` every CNAB invocation image runs, `/cnab/app/run` (Docker's
`--filter` flag doesn't support filtering by command, so filter client-side),
and take the first match:

```bash
docker ps -a --format '{{.ID}}\t{{.Command}}' | grep '/cnab/app/run' | head -n 1
```

Then commit it to an image and start a shell in it:

```bash
docker commit <container-id> porter-debug
docker run --rm -it --entrypoint bash porter-debug
```

### Access the Docker host's network from a bundle

A bundle's invocation image runs in its own container, so by default it can't resolve
hostnames or reach services that are only known to the Docker host, such as an entry in the
host's `/etc/hosts` or an internal DNS server that only the host (or CI runner) can reach.

Set `DOCKER_NETWORK=host` to run the bundle with the same network namespace as the Docker host,
giving it access to whatever hostnames and services the host itself can resolve:

```bash
DOCKER_NETWORK=host porter install --reference ghcr.io/getporter/examples/porter-hello:v0.2.0
```

⚠️️ Host networking is only supported on Linux Docker Engine. It does not work the same way on
Docker Desktop for Mac or Windows, since those run Docker inside a VM. See the
[troubleshooting guide][resolve-host-hostname] for alternatives that work everywhere.

[resolve-host-hostname]: /troubleshooting/#how-do-i-let-my-bundle-resolve-a-hostname-only-known-to-the-docker-host

## Next Steps

- [Connect to a Docker Registry](/operations/connect-registry/)
- [Configure Porter](/operations/configuration/)
