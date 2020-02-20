---
title: "porter mixins search"
slug: porter_mixins_search
url: /cli/porter_mixins_search/
---
## porter mixins search

Search available mixins

### Synopsis

Search available mixins. You can specify an optional mixin name query, where the results are filtered by mixins whose name contains the query term.

```
porter mixins search [QUERY] [flags]
```

### Examples

```
  porter mixin search
  porter mixin search helm
  porter mixin search -o json
```

### Options

```
  -h, --help            help for search
  -o, --output string   Output format, allowed values are: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter mixins](/cli/porter_mixins/)	 - Mixin commands. Mixins assist with authoring bundles.

