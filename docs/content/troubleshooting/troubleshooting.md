---
title: Troubleshooting
description: Error messages you may see from Porter and how to handle them
---

With any porter error, it can really help to re-run the command again with the `--debug` flag.

* [Examine Previous Logs](#examine-previous-logs)
* [Mapping values are not allowed in this context](#mapping-values-are-not-allowed-in-this-context)
* [You see apt errors when you use a custom Dockerfile](#you-see-apt-errors-when-you-use-a-custom-dockerfile)

## Examine Previous Logs

Porter [saves the logs](/operators/logs/) when a bundle is executed. Comparing the logs
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
have lines that contain a colon followed by a space `: ` or a hash `#` preceeded by a space, then
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
Starting Invocation Image Build =======>
Error: unable to build CNAB invocation image: failed to stream docker build output: The command '/bin/sh -c apt-get update && apt-get install -y apt-transport-https curl && curl -o kubectl https://storage.googleapis.com/kubernetes-release/release/v1.15.5/bin/linux/amd64/kubectl && mv kubectl /usr/local/bin && chmod a+x /usr/local/bin/kubectl' returned a non-zero code: 127
```

For now you must base your custom Dockerfile on debian or ubuntu.
