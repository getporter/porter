---
title: "porter credentials list"
slug: porter_credentials_list
url: /cli/porter_credentials_list/
---
## porter credentials list

List credentials

### Synopsis

List named sets of credentials defined by the user.

Optionally filters the results name, which returns all results whose name contain the provided query.
The results may also be filtered by associated labels and the namespace in which the credential set is defined.

```
porter credentials list [flags]
```

### Examples

```
  porter credentials list
  porter credentials list --namespace prod
  porter credentials list --all-namespaces,
  porter credentials list --name myapp
  porter credentials list --label env=dev
  porter credentials list --skip 2 --limit 2
```

### Options

```
      --all-namespaces     Include all namespaces in the results.
  -h, --help               help for list
  -l, --label strings      Filter the credential sets by a label formatted as: KEY=VALUE. May be specified multiple times.
      --limit int          Limit the number of credential sets by a certain amount. Defaults to 0.
      --name string        Filter the credential sets where the name contains the specified substring.
  -n, --namespace string   Namespace in which the credential set is defined. Defaults to the global namespace. Use * to list across all namespaces.
  -o, --output string      Specify an output format.  Allowed values: plaintext, json, yaml (default "plaintext")
      --skip int           Skip the number of credential sets by a certain amount. Defaults to 0.
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter credentials](/cli/porter_credentials/)	 - Credentials commands

