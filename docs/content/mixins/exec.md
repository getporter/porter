---
title: exec mixin
description: Run a command or script
---

Run a command or script.

✅ Learn how to use the exec mixin with our [Exec Mixin Best Practice Guide](/best-practices/exec-mixin/)

Source: https://github.com/deislabs/porter/tree/main/pkg/exec

### Install or Upgrade
```
porter mixin install exec
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
  suffix-arguments:
  - suffix-arg1
  suppress-output: false
  outputs:
  - name: NAME
    jsonPath: JSONPATH
  - name: NAME
    regex: GOLANG_REGULAR_EXPRESSION
  - name: NAME
    path: FILEPATH
```

This is executed as:

```
$ cmd arg1 arg2 -a flag-value --long-flag true --repeated-flag flag-value1 --repeated-flag flag-value2 suffix-arg1
```

### Suppress Output

The `suppress-output` field controls whether output from the mixin should be
prevented from printing to the console. By default this value is false, using
Porter's default behavior of hiding known sensitive values. When 
`suppress-output: true` all output from the mixin (stderr and stdout) are hidden.

Step outputs (below) are still collected when output is suppressed. This allows
you to prevent sensitive data from being exposed while still collecting it from
a command and using it in your bundle.

### Outputs

The mixin supports outputs of various types:

* [JSON Path](#json-path)
* [Regular Expressions](#regular-expressions)
* [File Paths](#file-paths)


#### JSON Path

The `jsonPath` output treats stdout like a json document and applies the expression, saving the result to the output.

```yaml
outputs:
- name: NAME
  jsonPath: JSONPATH
```

For example, if the `jsonPath` expression was `$[*].id` and the command sent the following to stdout: 

```json
[
  {
    "id": "1085517466897181794",
    "name": "my-vm"
  }
]
```

Then then output would have the following contents:

```json
["1085517466897181794"]
```

#### Regular Expressions

The `regex` output applies a Go-syntax regular expression to stdout and saves every capture group, one per line, to the output.

```yaml
outputs:
- name: NAME
  regex: GOLANG_REGULAR_EXPRESSION
```

For example, if the `regex` expression was `--- FAIL: (.*) \(.*\)` and the command send the following to stdout:

```
--- FAIL: TestMixin_Install (0.00s)
stuff
things
--- FAIL: TestMixin_Upgrade (0.00s)
more
logs
```

Then the output would have the following contents:

```
TestMixin_Install
TestMixin_Upgrade
```

#### File Paths

The `path` output saves the content of the specified file path to an output.

```yaml
outputs:
- name: kubeconfig
  path: /root/.kube/config
```

---

### Examples

See [exec outputs][exec-outputs] for a full working example.

Run a command
```yaml
install:
- exec:
    description: "Install Hello World"
    command: make
    arguments:
    - install
```

Run a script
```yaml
install:
- exec:
    description: "Install Hello World"
    command: ./install-world.sh
```

[exec-outputs]: https://porter.sh/src/examples/exec-outputs/

### FAQ

#### How do I use pipes?

If you have a command that pipes, place the command in a script file and then
use the exec mixin to invoke that script passing in any parameters that the
script requires as arguments.
