---
title: "porter installations output list"
slug: porter_installations_output_list
url: /cli/porter_installations_output_list/
---
## porter installations output list

List installation outputs

### Synopsis

Displays a listing of installation outputs.

```
porter installations output list [--installation|i INSTALLATION] [flags]
```

### Examples

```
  porter installation outputs list
    porter installation outputs list --installation another-bundle

```

### Options

```
  -h, --help                  help for list
  -i, --installation string   Specify the installation to which the output belongs.
  -o, --output string         Specify an output format.  Allowed values: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter installations output](/cli/porter_installations_output/)	 - Output commands

