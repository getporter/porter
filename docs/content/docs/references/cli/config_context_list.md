---
title: "porter config context list"
slug: porter_config_context_list
url: /cli/porter_config_context_list/
---
## porter config context list

List configuration contexts

### Synopsis

List all contexts defined in the porter configuration file. The active context is marked with *.

```
porter config context list [flags]
```

### Examples

```
  porter config context list
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter config context](/cli/porter_config_context/)	 - Context commands

