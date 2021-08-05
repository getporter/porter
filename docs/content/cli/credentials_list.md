---
title: "porter credentials list"
slug: porter_credentials_list
url: /cli/porter_credentials_list/
---
## porter credentials list

List credentials

### Synopsis

List named sets of credentials defined by the user.

```
porter credentials list [flags]
```

### Examples

```
  porter credentials list
  porter credentials list --namespace prod
  porter credentials list --namespace "*"
```

### Options

```
  -h, --help               help for list
  -n, --namespace string   Namespace in which the credential set is defined. Defaults to the global namespace. Use * to list across all namespaces.
  -o, --output string      Specify an output format.  Allowed values: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter credentials](/cli/porter_credentials/)	 - Credentials commands

