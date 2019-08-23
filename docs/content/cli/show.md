---
title: "porter show"
slug: porter_show
url: /cli/porter_show/
---
## porter show

Show an instance of a bundle

### Synopsis

Displays info relating to an instance of a bundle, including status and a listing of outputs.

```
porter show [INSTANCE] [flags]
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

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

