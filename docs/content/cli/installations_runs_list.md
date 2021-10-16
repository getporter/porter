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
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter installations runs](/cli/porter_installations_runs/)	 - Commands for working with runs of an Installation

