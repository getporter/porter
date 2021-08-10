---
title: Configuration
description: Controlling Porter with its config file, environment variables and flags
---

* [Flags](#flags)
* [Environment Variables](#environment-variables)
* [Config File](#config-file)
* [Experimental Feature Flags](#experimental-feature-flags)
  * [Build Drivers](#build-drivers)

Porter's configuration system has a precedence order:

* Flags (highest)
* Environment Variables
* Config File (lowest)

You may set a default value for a configuration value in the config file,
override it in a shell session with an environment variable and then override
both in a particular command with a flag.

* [Set Current Namespace](#namespace)
* [Enable Debug Output](#debug)
* [Debug Plugins](#debug-plugins)
* [Output Formatting](#output)
* [Allow Docker Host Access](#allow-docker-host-access)

## Flags

### Namespace
`--namespace` specifies the current namespace.

### Debug

`--debug` is a flag that is understood not only by the porter client but also the
runtime and most mixins. They may use it to print additional information that
may be useful when you think you may have found a bug, when you want to know
what commands they are executing, or when you need really verbose output to send
to the developers.

### Debug Plugins

`--debug-plugins` controls if logs related to communication
between porter and its plugins should be printed when debugging. This can be _very_
verbose, so it is not turned on by default when debug is true.

### Output

`--output` controls the format of the output printed by porter. Each command
supports a different set of allowed outputs though usually there is some
combination of: `table`, `json` and `yaml`.

### Allow Docker Host Access

`--allow-docker-host-access` controls whether the local Docker daemon
should be made available to executing bundles. This flag is available for the
following commands: [install], [upgrade], [invoke] and [uninstall]. When this
value is set to true, bundles are executed in a privileged container with the
docker socket mounted. This allows you to use Docker from within your bundle,
such as `docker push`, `docker-compose`, or docker-in-docker.

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

`--flag` has a corresponding environment variable of `PORTER_FLAG` and `--another-flag`
corresponds to the environment variable `PORTER_ANOTHER_FLAG`.

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
namespace = "dev"
debug = true
debug-plugins = true
output = "json"
allow-docker-host-access = true
```

## Experimental Feature Flags

Porter sometimes uses feature flags to release new functionality for users to
evaluate, without affecting the stability of Porter. You can enable an experimental
feature by:

* Using the experimental global flag `--experimental flagA,flagB`.
  The value is a comma-separated list of strings.
* Setting the PORTER_EXPERIMENTAL environment variable like so `PORTER_EXPERIMENTAL=flagA,flagB`.
  The value is a comma-separated list of strings.
* Setting the experimental field in the configuration file like so `experimental = ["flagA","flagB"]`.
  The value is an array of strings.

### Build Drivers

The **build-drivers** experimental feature flag enables using a different
driver to build OCI images used by the bundle, such as the installer.

You can set your desired driver with either using `porter build --driver`,
`PORTER_BUILD_DRIVER` environment variable, or in the configuration file with
`build-driver = "DRIVER"`

The default driver is [Docker], and the full list of available drivers
is below.

* **Docker**: Build an OCI image using the [Docker library], without buildkit support.
  This requires access to a Docker daemon, either locally or remote.
* **Buildkit**: Build an OCI image using [Docker with Buildkit].
  With buildkit you can improve the performance of builds using caching, access
  private resources during build, and more. 
  This requires access to a Docker daemon, either locally or remote.

Below are some examples of how to enable the build-drivers feature and specify an alternate
driver:

**Flags**
```
porter build --experimental build-drivers --driver buildkit
```

**Environment Variables**
```
export PORTER_EXPERIMENTAL=build-drivers
export PORTER_BUILD_DRIVER=buildkit
```

**Configuration File**
```toml
experimental = ["build-drivers"]
build-driver = "buildkit"
```

[install]: /cli/porter_install/
[upgrade]: /cli/porter_upgrade/
[invoke]: /cli/porter_invoke/
[uninstall]: /cli/porter_uninstall/
[Docker library]: https://github.com/moby/moby
[Docker with Buildkit]: https://docs.docker.com/develop/develop-images/build_enhancements/
