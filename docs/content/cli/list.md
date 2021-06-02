---
title: "porter list"
slug: porter_list
url: /cli/porter_list/
---
## porter list

List installed bundles

### Synopsis

List all bundles installed by Porter.

A listing of bundles currently installed by Porter will be provided, along with metadata such as creation time, last action, last status, etc.

Optional output formats include json and yaml.

```
porter list [flags]
```

### Examples

```
  porter list
  porter list -o json
```

### Options

```
  -h, --help            help for list
  -o, --output string   Specify an output format.  Allowed values: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

