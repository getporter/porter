---
title: "porter config show"
slug: porter_config_show
url: /cli/porter_config_show/
---
## porter config show

Show Porter configuration

### Synopsis

Display the current Porter configuration. If no config file exists, shows default values.

```
porter config show [flags]
```

### Examples

```
  porter config show
  porter config show -o json
  porter config show -o yaml
  porter config show -o toml
```

### Options

```
  -h, --help            help for show
  -o, --output string   Output format (json, yaml, toml)
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter config](/cli/porter_config/)	 - Config commands

