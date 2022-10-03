---
title: exec mixin
description: Run a command or script
---

Run a command or script.

✅ Learn how to use the exec mixin with our [Exec Mixin Best Practice Guide](/best-practices/exec-mixin/)

Source: https://getporter.org/src/pkg/exec

### Install or Upgrade
```
porter mixin install exec --version v1.0.0-rc.3
```

## Mixin Syntax

```yaml
exec:
  description: "Description of the command"
  command: cmd # The command to run, must be on the PATH
  arguments: # arguments to pass to the command
  - arg1
  - arg2
  flags: # flags to pass to the command, porter determines if it is a long (--flag) or short flag (-f)
    a: flag-value
    long-flag: true
    repeated-flag: # Use an array if a flag must be specified multiple times with different values
    - flag-value1
    - flag-value2
  suffix-arguments: # These arguments are specified after any flags are passed
  - suffix-arg1
  envs: # Environment variables to be added to the command execution environment
    FOO_KEY: foo-value
  suppress-output: false # Do not print the command output to the console
  ignoreError: # Conditions when execution should continue even if the command fails
    all: true # Ignore all errors 
    exitCodes: # Ignore failed commands that return the following exit codes
      - 1
      - 2
    output: # Ignore failed commands based on the contents of stderr
      contains: # Ignore when stderr contains a substring
        - "SUBSTRING IN STDERR"
      regex: # Ignore when stderr matches a regular expression
        - "GOLANG_REGULAR_EXPRESSION"
  outputs: # Collect values from the command and make it available as an output
  - name: NAME
    jsonPath: JSONPATH # Scrape stdout with a json path expression
  - name: NAME
    regex: GOLANG_REGULAR_EXPRESSION # Scrape stdout with a regular expression
  - name: NAME
    path: FILEPATH # Save the contents of a file
```

This is executed as:

```
$ cmd arg1 arg2 -a flag-value --long-flag true --repeated-flag flag-value1 --repeated-flag flag-value2 suffix-arg1
```

### Suppress Output

The `suppress-output` field controls whether output from the mixin should be
prevented from printing to the console. By default, this value is false, using
Porter's default behavior of hiding known sensitive values. When 
`suppress-output: true` all output from the mixin (stderr and stdout) are hidden.

Step outputs (below) are still collected when output is suppressed. This allows
you to prevent sensitive data from being exposed while still collecting it from
a command and using it in your bundle.

### Ignore Error

In some cases, you may need to have the bundle continue executing when a mixin command fails.
For example when the command fails because the resource already exists.

You can ignore errors based on:

* All - Ignore all errors from the command.
* ExitCodes - Ignore errors when one of the specified exit codes are returned.
* Output Contains - Ignore errors when the command's stderr contains the specified string.
* Output Regex - Ignore errors when the command's stderr matches the specified regular expression (in Go syntax).

Porter only prints out that an error was ignored in debug mode.

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

Then the output would have the following contents:

```json
["1085517466897181794"]
```

When you are developing your jsonPath expression, you can specify the --debug
flag, and the full json document with your query are printed to stderr so that you
can troubleshoot and improve your query based on the real result of the mixin's
execution.

Note: Porter attempts to preserve the original format of numeric values, so if the value
is in scientific notation, the captured output should also be in scientific notation.
Conversely, if the original number was _not_ in scientific notation, then the captured
value should also not be in scientific notation. Please open a bug if you find that the
format doesn't match what you expected.

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
  path: /home/nonroot/.kube/config
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

[exec-outputs]: /examples/src/exec-outputs/

### FAQ

#### How do I use pipes?

If you have a command that pipes, place the command in a script file and then
use the exec mixin to invoke that script passing in any parameters that the
script requires as arguments.
