---
title: "porter instances output list"
slug: porter_instances_output_list
url: /cli/porter_instances_output_list/
---
## porter instances output list

List bundle instance outputs

### Synopsis

Displays a listing of bundle instance outputs.

```
porter instances output list [--instance|i INSTANCE] [flags]
```

### Examples

```
  porter instance outputs list
    porter instance outputs list --instance another-bundle

```

### Options

```
  -h, --help              help for list
  -i, --instance string   Specify the bundle instance to which the output belongs.
  -o, --output string     Specify an output format.  Allowed values: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter instances output](/cli/porter_instances_output/)	 - Output commands

