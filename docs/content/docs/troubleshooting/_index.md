---
title: Troubleshooting
description: Error messages you may see from Porter and how to handle them
weight: 10
---

With any porter error, it can really help to re-run the command again with the `--debug` flag.

- [Examine Previous Logs](#examine-previous-logs)
- [Mapping values are not allowed in this context](#mapping-values-are-not-allowed-in-this-context)
- [You see apt errors when you use a custom Dockerfile](#you-see-apt-errors-when-you-use-a-custom-dockerfile)
- [I want to inspect the container after a bundle action runs](#i-want-to-inspect-the-container-after-a-bundle-action-runs)
- [How do I let my bundle resolve a hostname only known to the Docker host?](#how-do-i-let-my-bundle-resolve-a-hostname-only-known-to-the-docker-host)

## Examine Previous Logs

Porter [saves the logs](/docs/operations/view-logs/) when a bundle is executed. Comparing the logs
from a failing run to those from a successful run may assist with
troubleshooting.

## Mapping values are not allowed in this context

When you run your bundle you see the following error

```
executing install action from porter-hello (installation: porter-hello)
=== Step Data ===
map[bundle:map[credentials:map[] dependencies:map[] description:An example Porter configuration images:map[] invocationImage:porter-hello:latest name:porter-hello outputs:map[] parameters:map[test:{"test": "test"}] version:0.1.0]]
=== Step Template ===
exec:
  command: bash
  description: Install Hello World
  flags:
    c: echo '${ bundle.parameters.test }'

=== Rendered Step ===
exec:
  command: bash
  description: Install Hello World
  flags:
    c: echo '{"test": "value"}'

Error: unable to resolve step: invalid step yaml
exec:
  command: bash
  description: Install Hello World
  flags:
    c: echo '{"test": "value"}'
: yaml: line 5: mapping values are not allowed in this context
```

Right now Porter [doesn't preserve the wrapping quotes around mapping values][851], so if you
have lines that contain a colon followed by a space `: ` or a hash `#` preceded by a space, then
things will get tricky. If you can remove the space, or wrap the entire line in an extra quote, that
should workaround the problem.

[851]: https://github.com/getporter/porter/issues/851

**before**

```yaml
parameters:
- name: test
  description: test
  type: string
  default: '{"test": "value"}'

install:
  - exec:
      description: "Install Hello World"
      command: bash
      flags:
        c: "echo ${ bundle.parameters.test}
```

**after**

Remove the extra space after the colon when defining the test parameter's default

```yaml
parameters:
- name: test
  description: test
  type: string
  default: '{"test":"value"}'

install:
  - exec:
      description: "Install Hello World"
      command: bash
      flags:
        c: "echo ${ bundle.parameters.test}
```

## You see apt errors when you use a custom Dockerfile

When you use a custom Dockerfile you see `apt` errors even though you did not use apt in your Dockerfile. This is because
Porter assumes a debian-based base image that has apt available. Many of the mixins use apt to install the dependencies
and the binary that they shim.

```
Starting Bundle Image Build =======>
Error: unable to build CNAB bundle image: failed to stream docker build output: The command '/bin/sh -c apt-get update && apt-get install -y apt-transport-https curl && curl -o kubectl https://storage.googleapis.com/kubernetes-release/release/v1.15.5/bin/linux/amd64/kubectl && mv kubectl /usr/local/bin && chmod a+x /usr/local/bin/kubectl' returned a non-zero code: 127
```

For now you must base your custom Dockerfile on debian or ubuntu.

## I want to inspect the container after a bundle action runs

Porter removes the container it used to run a bundle action once the action
finishes. To leave the container behind so you can inspect its filesystem,
see [Inspect the container after a bundle action runs][cleanup-containers].

[cleanup-containers]: /operations/connect-to-docker/#inspect-the-container-after-a-bundle-action-runs

## How do I let my bundle resolve a hostname only known to the Docker host?

A bundle's invocation image runs in its own container, so it can't resolve hostnames or reach
services that only the Docker host (or CI runner) knows about, such as a name mapped in the
host's `/etc/hosts` or resolvable only through the host's internal DNS.

**On Linux**, run the bundle with [`DOCKER_NETWORK=host`][access-host-network] so it shares the
host's network namespace and can resolve anything the host can.

**On Docker Desktop for Mac or Windows**, use `host.docker.internal` inside the bundle to reach
the Docker host's own IP address directly. This isn't available on Linux Docker Engine.

You can't work around this by writing custom entries into the bundle's own `/etc/hosts`:

- A `file`-type parameter (see [File Parameters][file-parameters]) can't be mapped directly to
  `/etc/hosts` — Docker manages that path as a special bind-mounted file, so Porter can't
  pre-populate it that way, and the bundle fails to start with `unable to decode parameter:
  illegal base64 data`.
- Appending to `/etc/hosts` yourself in an install step (`echo ... >> /etc/hosts`) requires the
  container to run as root, which isn't possible either: Porter always runs the bundle's
  invocation image as a non-root user, even with a [custom Dockerfile][custom-dockerfile] — it's
  not something a bundle can opt out of.

[access-host-network]: /operations/connect-to-docker/#access-the-docker-hosts-network-from-a-bundle
[file-parameters]: /bundle/manifest/#file-parameters
[custom-dockerfile]: /bundle/custom-dockerfile/
