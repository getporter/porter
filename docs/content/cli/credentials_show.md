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
  porter credential show NAME [-o table|json|yaml]
```

### Options

```
  -h, --help            help for show
  -o, --output string   Specify an output format.  Allowed values: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter credentials](/cli/porter_credentials/)	 - Credentials commands

