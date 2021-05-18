---
title: "Example: Hello World"
---

Source: https://porter.sh/src/examples/hello

The hello world bundle, [getporter/porter-hello], is the most simple bundle possible.
It prints a message to the console when various actions are performed.
This is the default bundle generated for you when you run `porter create`.

## Try it out

1. Use `porter explain` to see what is included in the bundle and how to use it.
    ```bash
    porter explain --reference getporter/porter-hello:v0.1.1
    ```

1. Install the bundle
    ```bash
    porter install hello --reference getporter/porter-hello:v0.1.1
    ```

1. Upgrade the bundle
    ```bash
    porter upgrade hello
    ```

1. Uninstall the bundles
    ```bash
    porter uninstall hello
    ```

[getporter/porter-hello]: https://hub.docker.com/r/getporter/porter-hello/
