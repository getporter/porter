---
title: Configuration
description: Controlling Porter with its config file, environment variables and flags
---

Porter's configuration system has a precedence order:

* Flags (highest)
* Environment Variables
* Config File (lowest)

You may set a default value for a configuration value in the config file,
override it in a shell session with an environment variable and then override
both in a particular command with a flag.

## Flags

### Debug

`--debug` is a flag that is understood not only by the porter client but also the
runtime and most mixins. They may use it to print additional information that
may be useful when you think you may have found a bug, when you want to know
what commands they are executing, or when you need really verbose output to send
to the developers.

### Output

`--output` controls the format of the output printed by porter. Each command
supports a different set of allowed outputs though usually there is some
combination of: `table`, `json` and `yaml`.

## Environment Variables

Flags have corresponding environment variables that you can use so that you
don't need to manually set the flag every time. The flag will default to the
value of the environment variable, when defined.

`--flag` has a corresponding environment variable of `PORTER_FLAG`

For example, you can set `PORTER_DEBUG=true` and then all subsequent porter
commands will act as though the `--debug` flag was passed.

## Config File

Common flags can be defaulted in the config file. The config file is located in
the PORTER_HOME directory (**~/.porter**), is named **config** and can be in any
of the following file types: JSON, TOML, YAML, HCL, envfile and Java Properties
files.

Below is an example configuration file in TOML

**~/.porter/config.toml**
```toml
debug = true
output = "json"
```
