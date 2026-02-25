---
title: "porter config edit"
slug: porter_config_edit
url: /cli/porter_config_edit/
---
## porter config edit

Edit the config file

### Synopsis

Edit the porter configuration file.
If the config file does not exist, a default template will be created.

Uses the EDITOR environment variable to determine which editor to use.

```
porter config edit [flags]
```

### Examples

```
  porter config edit
```

### Options

```
  -h, --help   help for edit
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter config](/cli/porter_config/)	 - Config commands

