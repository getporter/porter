---
title: "porter installations runs list"
slug: porter_installations_runs_list
url: /cli/porter_installations_runs_list/
---
## porter installations runs list

List runs of an Installation

### Synopsis

List runs of an Installation

```
porter installations runs list [flags]
```

### Examples

```
  porter installation runs list [NAME] [--namespace NAMESPACE] [--output FORMAT]

  porter installations runs list --name myapp --namespace dev


```

### Options

```
  -h, --help               help for list
  -n, --namespace string   Namespace in which the installation is defined. Defaults to the global namespace.
  -o, --output string      Specify an output format.  Allowed values: plaintext, json, yaml (default "plaintext")
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter installations runs](/cli/porter_installations_runs/)	 - Commands for working with runs of an Installation

