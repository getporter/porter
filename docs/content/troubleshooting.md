---
title: Troubleshooting
description: Error messages you may see from Porter and how to handle them
---

With any porter error, it can really help to re-run the command again with the `--debug` flag.

* [mapping values are not allowed in this context](#mapping-values-are-not-allowed-in-this-context)

## mapping values are not allowed in this context

When you run your bundle you see the following error

```
executing install action from HELLO (bundle instance: HELLO)
=== Step Data ===
map[bundle:map[credentials:map[] dependencies:map[] description:An example Porter configuration images:map[] invocationImage:porter-hello:latest name:HELLO outputs:map[] parameters:map[test:{"test": "test"}] version:0.1.0]]
=== Step Template ===
exec:
  command: bash
  description: Install Hello World
  flags:
    c: echo '{{ bundle.parameters.test }}'

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

[851]: https://github.com/deislabs/porter/issues/851
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
        c: "echo {{ bundle.parameters.test}}"
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
        c: "echo {{ bundle.parameters.test}}"
```