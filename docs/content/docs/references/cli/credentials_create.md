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
  porter credentials create mycreds.yaml --reference ghcr.io/getporter/examples/credentials-tutorial:v0.3.0
```

### Options

```
      --force               Force a fresh pull of the bundle
  -h, --help                help for create
      --insecure-registry   Don't require TLS for the registry
      --output string       Credential set resource file format
  -r, --reference string    Create the credential set from the credentials defined by the bundle at the specified OCI reference, e.g. ghcr.io/getporter/examples/porter-hello:v0.2.0
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter credentials](/cli/porter_credentials/)	 - Credentials commands

