---
title: "porter credentials show"
slug: porter_credentials_show
url: /cli/porter_credentials_show/
---
## porter credentials show

Show a Credential

### Synopsis

Show a particular credential set, including all named credentials and their corresponding mappings.

```
porter credentials show [flags]
```

### Examples

```
  porter credential show github --namespace dev
  porter credential show prodcluster --output json
```

### Options

```
  -h, --help               help for show
  -n, --namespace string   Namespace in which the credential set is defined. Defaults to the global namespace.
  -o, --output string      Specify an output format.  Allowed values: plaintext, json, yaml (default "plaintext")
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter credentials](/cli/porter_credentials/)	 - Credentials commands

