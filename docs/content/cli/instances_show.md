---
title: "porter instances show"
slug: porter_instances_show
url: /cli/porter_instances_show/
---
## porter instances show

Show an instance of a bundle

### Synopsis

Displays info relating to an instance of a bundle, including status and a listing of outputs.

```
porter instances show [INSTANCE] [flags]
```

### Examples

```
  porter instance show
porter instance show another-bundle

Optional output formats include json and yaml.

```

### Options

```
  -h, --help            help for show
  -o, --output string   Specify an output format.  Allowed values: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter instances](/cli/porter_instances/)	 - Bundle Instance commands

