---
title: "porter parameters create"
slug: porter_parameters_create
url: /cli/porter_parameters_create/
---
## porter parameters create

Create a Parameter Set

### Synopsis

Create a new blank resource for the definition of a Parameter Set.

```
porter parameters create [flags]
```

### Examples

```

		porter parameters create FILE [--output yaml|json]
		porter parameters create parameter-set.json
		porter parameters create parameter-set --output yaml
		porter parameters create myparams.yaml --reference ghcr.io/getporter/examples/porter-hello:v0.2.0
```

### Options

```
      --force               Force a fresh pull of the bundle
  -h, --help                help for create
      --insecure-registry   Don't require TLS for the registry
      --output string       Parameter set resource file format
  -r, --reference string    Create the parameter set from the parameters defined by the bundle at the specified OCI reference, e.g. ghcr.io/getporter/examples/porter-hello:v0.2.0
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter parameters](/cli/porter_parameters/)	 - Parameter set commands

