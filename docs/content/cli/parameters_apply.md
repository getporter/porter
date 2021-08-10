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
  porter parameters apply --file myparams.yaml
```

### Options

```
  -h, --help               help for apply
  -n, --namespace string   Namespace in which the parameter set is defined. The namespace in the file, if set, takes precedence.
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter parameters](/cli/porter_parameters/)	 - Parameter set commands

