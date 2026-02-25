---
title: "porter config"
slug: porter_config
url: /cli/porter_config/
---
## porter config

Config commands

### Synopsis

Commands for managing Porter's configuration file.

### Options

```
  -h, --help   help for config
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter](/cli/porter/)	 - With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://porter.sh/quickstart to learn how to use Porter.

* [porter config context](/cli/porter_config_context/)	 - Context commands
* [porter config edit](/cli/porter_config_edit/)	 - Edit the config file
* [porter config migrate](/cli/porter_config_migrate/)	 - Migrate the config file to the multi-context format
* [porter config show](/cli/porter_config_show/)	 - Show the config file

