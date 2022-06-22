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
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter parameters](/cli/porter_parameters/)	 - Parameter set commands

