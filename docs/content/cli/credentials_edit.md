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
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter credentials](/cli/porter_credentials/)	 - Credentials commands

