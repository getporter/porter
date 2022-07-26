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
```

### Options

```
  -h, --help            help for create
      --output string   Parameter set resource file format
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter parameters](/cli/porter_parameters/)	 - Parameter set commands

