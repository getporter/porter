---
title: Porter Commands
description: Porter CLI Commands Explained
---


# Init

Initialize the porter home directory (`~/.porter`).

```console
$ porter init
TODO: NOT IMPLEMENTED
```

# Create

Scaffold a new porter bundle in the current directory.

```console
$ porter create ...
TODO: INSERT CREATE OUTPUT
```

# Build

Build the bundle in the current directory.

```console
$ porter build
TODO: INSERT BUILD OUTPUT
```

# Run

This is a runtime command and isn't a command that you should run yourself. TODO: Move into a separate section on the runtime commands.

```console
$ porter run ...
TODO: INSERT RUN OUTPUT
```

# List Mixins

List the mixins installed in the `PORTER_HOME/mixins` directory.

```console
$ porter list mixins
exec
helm
```

**Flags**

* `--output`, `-o`: Output format, allowed values are: table, json (default "table")

# Completion

Generate a bash completion script for porter.

```console
$ porter completion
TODO: NOT IMPLEMENTED
```

# Version

Print the porter cli version to the console.

```console
$ porter version
```