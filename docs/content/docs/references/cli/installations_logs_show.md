---
title: "porter installations logs show"
slug: porter_installations_logs_show
url: /cli/porter_installations_logs_show/
---
## porter installations logs show

Show the logs from an installation

### Synopsis

Show the logs from an installation.

Either display the logs from a specific run of a bundle with --run, or use --installation to display the logs from its most recent run.

```
porter installations logs show [flags]
```

### Examples

```
  porter installation logs show --installation wordpress --namespace dev
  porter installations logs show --run 01EZSWJXFATDE24XDHS5D5PWK6
```

### Options

```
  -h, --help                  help for show
  -i, --installation string   The installation that generated the logs.
  -n, --namespace string      Namespace in which the installation is defined. Defaults to the global namespace.
  -r, --run string            The bundle run that generated the logs.
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter installations logs](/cli/porter_installations_logs/)	 - Installation Logs commands

