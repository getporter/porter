---
title: "porter parameters list"
slug: porter_parameters_list
url: /cli/porter_parameters_list/
---
## porter parameters list

List parameter sets

### Synopsis

List named sets of parameters defined by the user.

```
porter parameters list [flags]
```

### Examples

```
  porter parameters list
  porter parameters list --namespace prod -o json
  porter parameters list --namespace "*"
```

### Options

```
  -h, --help               help for list
  -n, --namespace string   Namespace in which the parameter set is defined. Defaults to the global namespace. Use * to list across all namespaces.
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

