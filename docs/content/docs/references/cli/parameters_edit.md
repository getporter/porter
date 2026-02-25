---
title: "porter parameters edit"
slug: porter_parameters_edit
url: /cli/porter_parameters_edit/
---
## porter parameters edit

Edit Parameter Set

### Synopsis

Edit a named parameter set.

```
porter parameters edit [flags]
```

### Examples

```
  porter parameter edit debug-tweaks --namespace dev
```

### Options

```
  -h, --help               help for edit
  -n, --namespace string   Namespace in which the parameter set is defined. Defaults to the global namespace.
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter parameters](/cli/porter_parameters/)	 - Parameter set commands

