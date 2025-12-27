---
title: "porter config set"
slug: porter_config_set
url: /cli/porter_config_set/
---
## porter config set

Set a config value

### Synopsis

Set an individual Porter configuration value. Creates a config file if none exists.

```
porter config set KEY VALUE [flags]
```

### Examples

```
  porter config set verbosity debug
  porter config set logs.level info
  porter config set telemetry.enabled true
  porter config set namespace myapp
```

### Options

```
  -h, --help   help for set
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter config](/cli/porter_config/)	 - Config commands

