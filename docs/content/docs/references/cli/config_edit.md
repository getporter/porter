---
title: "porter config edit"
slug: porter_config_edit
url: /cli/porter_config_edit/
---
## porter config edit

Edit Porter configuration

### Synopsis

Edit the Porter configuration in your default editor. If no config file exists, creates a default configuration file.

```
porter config edit [flags]
```

### Examples

```
  porter config edit
  EDITOR=vim porter config edit
```

### Options

```
  -h, --help   help for edit
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter config](/cli/porter_config/)	 - Config commands

