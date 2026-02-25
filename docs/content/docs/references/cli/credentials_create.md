---
title: "porter credentials create"
slug: porter_credentials_create
url: /cli/porter_credentials_create/
---
## porter credentials create

Create a Credential

### Synopsis

Create a new blank resource for the definition of a Credential Set.

```
porter credentials create [flags]
```

### Examples

```
  porter credentials create FILE [--output yaml|json]
  porter credentials create credential-set.json
  porter credentials create credential-set --output yaml
```

### Options

```
  -h, --help            help for create
      --output string   Credential set resource file format
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter credentials](/cli/porter_credentials/)	 - Credentials commands

