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
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter credentials](/cli/porter_credentials/)	 - Credentials commands

