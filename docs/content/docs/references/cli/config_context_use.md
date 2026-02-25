---
title: "porter config context use"
slug: porter_config_context_use
url: /cli/porter_config_context_use/
---
## porter config context use

Set the current configuration context

### Synopsis

Set the current-context in the porter configuration file.

```
porter config context use <name> [flags]
```

### Examples

```
  porter config context use prod
```

### Options

```
  -h, --help   help for use
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter config context](/cli/porter_config_context/)	 - Context commands

