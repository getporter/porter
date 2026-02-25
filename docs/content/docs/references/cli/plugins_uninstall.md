---
title: "porter plugins uninstall"
slug: porter_plugins_uninstall
url: /cli/porter_plugins_uninstall/
---
## porter plugins uninstall

Uninstall a plugin

```
porter plugins uninstall NAME [flags]
```

### Examples

```
  porter plugin uninstall azure
```

### Options

```
  -h, --help   help for uninstall
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter plugins](/cli/porter_plugins/)	 - Plugin commands. Plugins enable Porter to work on different cloud providers and systems.

