---
title: "porter installations output list"
slug: porter_installations_output_list
url: /cli/porter_installations_output_list/
---
## porter installations output list

List installation outputs

### Synopsis

Displays a listing of installation outputs.

```
porter installations output list [--installation|i INSTALLATION] [flags]
```

### Examples

```
  porter installation outputs list
    porter installation outputs list --installation another-bundle
    porter installation outputs list --run 01EZSWJXFATDE24XDHS5D5PWK6

```

### Options

```
  -h, --help                  help for list
  -i, --installation string   Specify the installation to which the output belongs.
  -n, --namespace string      Namespace in which the installation is defined. Defaults to the global namespace.
  -o, --output string         Specify an output format.  Allowed values: plaintext, json, yaml (default "plaintext")
  -r, --run string            The bundle run that generated the outputs.
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter installations output](/cli/porter_installations_output/)	 - Output commands

