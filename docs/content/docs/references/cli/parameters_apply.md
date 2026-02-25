---
title: "porter parameters apply"
slug: porter_parameters_apply
url: /cli/porter_parameters_apply/
---
## porter parameters apply

Apply changes to a parameter set

### Synopsis

Apply changes from the specified file to a parameter set. If the parameter set doesn't already exist, it is created.

Supported file extensions: json and yaml.

You can use the generate and show commands to create the initial file:
  porter parameters generate myparams --reference SOME_BUNDLE
  porter parameters show myparams --output yaml > myparams.yaml


```
porter parameters apply FILE [flags]
```

### Examples

```
  porter parameters apply myparams.yaml
```

### Options

```
  -h, --help               help for apply
  -n, --namespace string   Namespace in which the parameter set is defined. The namespace in the file, if set, takes precedence.
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter parameters](/cli/porter_parameters/)	 - Parameter set commands

