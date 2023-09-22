---
title: "porter credentials edit"
slug: porter_credentials_edit
url: /cli/porter_credentials_edit/
---
## porter credentials edit

Edit Credential

### Synopsis

Edit a named credential set.

```
porter credentials edit [flags]
```

### Examples

```
  porter credentials edit github --namespace dev
```

### Options

```
  -h, --help               help for edit
  -n, --namespace string   Namespace in which the credential set is defined. Defaults to the global namespace.
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter credentials](/cli/porter_credentials/)	 - Credentials commands

