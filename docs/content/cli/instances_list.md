---
title: "porter instances list"
slug: porter_instances_list
url: /cli/porter_instances_list/
---
## porter instances list

List instances of installed bundles

### Synopsis

List instances of all bundles installed by Porter.

A listing of instances of bundles currently installed by Porter will be provided, along with metadata such as creation time, last action, last status, etc.

Optional output formats include json and yaml.

```
porter instances list [flags]
```

### Examples

```
  porter instances list
  porter instances list -o json
```

### Options

```
  -h, --help            help for list
  -o, --output string   Specify an output format.  Allowed values: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter instances](/cli/porter_instances/)	 - Bundle Instance commands

