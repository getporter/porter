---
title: "porter parameters show"
slug: porter_parameters_show
url: /cli/porter_parameters_show/
---
## porter parameters show

Show a Parameter Set

### Synopsis

Show a named parameter set, including all named parameters and their corresponding mappings.

```
porter parameters show [flags]
```

### Examples

```
  porter parameter show NAME [-o table|json|yaml]
```

### Options

```
  -h, --help               help for show
  -n, --namespace string   Namespace in which the parameter set is defined. Defaults to the global namespace.
  -o, --output string      Specify an output format.  Allowed values: plaintext, json, yaml (default "plaintext")
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter parameters](/cli/porter_parameters/)	 - Parameter set commands

