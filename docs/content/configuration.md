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

* [Enable Debug Output](#debug)
* [Output Formatting](#output)
* [Allow Docker Host Access](#allow-docker-host-access)

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

### Allow Docker Host Access

`--allow-docker-host-access` controls whether or not the local Docker daemon
should be made available to executing bundles. This flag is available for the
following commands: [install], [upgrade], [invoke] and [uninstall]. When this
value is set to true, bundles will be executed with the docker socket mounted.
This allows you to use Docker from within your bundle, such as `docker push`,
`docker-compose`, or docker-in-docker. In addition, configuration may include
running the container in privileged mode and/or with host networking enabled. 

🚨 **There are security implications to enabling access! You should trust any
bundles that you execute with this setting enabled as it gives them elevated 
access to the host machine.**

⚠️️ This configuration setting is only available when you are in an environment 
that provides access to the local docker daemon. Therefore it does not work with
the Azure Cloud Shell driver.

## Environment Variables

Flags have corresponding environment variables that you can use so that you
don't need to manually set the flag every time. The flag will default to the
value of the environment variable, when defined.

`--flag` has a corresponding environment variable of `PORTER_FLAG`

For example, you can set `PORTER_DEBUG=true` and then all subsequent porter
commands will act as though the `--debug` flag was passed.

## Config File

Common settings can be defaulted in the config file. The config file is located in
the PORTER_HOME directory (**~/.porter**), is named **config** and can be in any
of the following file types: JSON, TOML, YAML, HCL, envfile and Java Properties
files.

Below is an example configuration file in TOML

**~/.porter/config.toml**
```toml
debug = true
output = "json"
allow-docker-host-access = true
```

[install]: /cli/porter_install/
[upgrade]: /cli/porter_upgrade/
[invoke]: /cli/porter_invoke/
[uninstall]: /cli/porter_uninstall/