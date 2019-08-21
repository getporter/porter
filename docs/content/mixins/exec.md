---
title: exec mixin
description: Using the exec mixin
---

Run a command or script

Source: https://github.com/deislabs/porter/tree/master/pkg/exec

### Install or Upgrade
```
porter mixin install terraform --feed-url https://cdn.deislabs.io/porter/atom.xml
```

## Mixin Syntax

```yaml
exec:
  description: "Description of the command"
  command: cmd
  arguments:
  - arg1
  - arg2
  flags:
    a: flag-value
    long-flag: true
    repeated-flag:
    - flag-value1
    - flag-value2
```

This is executed as:

```
$ cmd arg1 arg2 -a flag-value --long-flag true --repeated-flag flag-value1 --repeated-flag flag-value2
```

### Examples

Run a command
```yaml
install:
- exec:
    description: "Install Hello World"
    command: bash
    flags:
      c: echo Hello World
```

Run a script
```yaml
install:
- exec:
    description: "Install Hello World"
    command: bash
    arguments:
    - ./install-world.sh
```

### FAQ

#### How do I use pipes?

If you have a command that pipes, place the command in a script file and then
use the exec mixin to invoke that script passing in any parameters that the
script requires as arguments.
