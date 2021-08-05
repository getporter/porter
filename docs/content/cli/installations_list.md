---
title: "porter installations list"
slug: porter_installations_list
url: /cli/porter_installations_list/
---
## porter installations list

List installed bundles

### Synopsis

List all bundles installed by Porter.

A listing of bundles currently installed by Porter will be provided, along with metadata such as creation time, last action, last status, etc.

Optional output formats include json and yaml.

```
porter installations list [flags]
```

### Examples

```
  porter installations list
  porter installations list -o json
```

### Options

```
  -h, --help               help for list
  -n, --namespace string   Filter the installations by namespace. Defaults to the global namespace.
  -o, --output string      Specify an output format.  Allowed values: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter installations](/cli/porter_installations/)	 - Installation commands

