---
title: 'porter credentials delete'
aliases:
  - /cli/porter_credentials_delete/
---

## porter credentials delete

Delete a Credential

### Synopsis

Delete a named credential set.

```
porter credentials delete NAME [flags]
```

### Examples

```
  porter credentials delete github --namespace dev
```

### Options

```
  -h, --help               help for delete
  -n, --namespace string   Namespace in which the credential set is defined. Defaults to the global namespace.
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter credentials](/cli/porter_credentials/)	 - Credentials commands

