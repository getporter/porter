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
  -o, --output string      Specify an output format.  Allowed values: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter parameters](/cli/porter_parameters/)	 - Parameter set commands

