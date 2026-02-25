---
title: "porter plugins show"
slug: porter_plugins_show
url: /cli/porter_plugins_show/
---
## porter plugins show

Show details about an installed plugin

```
porter plugins show [flags]
```

### Options

```
  -h, --help            help for show
  -o, --output string   Output format, allowed values are: plaintext, json, yaml (default "plaintext")
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter plugins](/cli/porter_plugins/)	 - Plugin commands. Plugins enable Porter to work on different cloud providers and systems.

