---
title: "Example: Hello World"
description: "Learn how to install a bundle"
weight: 0
---

Source: https://getporter.org/examples/src/hello

The hello world bundle, [ghcr.io/getporter/examples/porter-hello], is the simplest bundle possible.
It prints a message to the console when various actions are performed.
This is the default bundle generated for you when you run `porter create`.

## Try it out

1. Use `porter explain` to see what is included in the bundle and how to use it.
    ```console
    porter explain ghcr.io/getporter/examples/porter-hello:v0.2.0
    ```

1. Install the bundle
    ```
    porter install hello --reference ghcr.io/getporter/examples/porter-hello:v0.2.0
    ```

1. Upgrade the bundle
    ```
    porter upgrade hello
    ```

1. Uninstall the bundle
    ```
    porter uninstall hello
    ```


[ghcr.io/getporter/examples/porter-hello]: https://github.com/getporter/examples/pkgs/container/examples%2Fporter-hello
