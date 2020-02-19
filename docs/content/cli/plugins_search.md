---
title: "porter plugins search"
slug: porter_plugins_search
url: /cli/porter_plugins_search/
---
## porter plugins search

Search available plugins

### Synopsis

Search available plugins

```
porter plugins search [NAME] [flags]
```

### Examples

```
  porter plugin search
  porter plugin search azure
  porter plugin search azure -o json
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

* [porter plugins](/cli/porter_plugins/)	 - Plugin commands. Plugins enable Porter to work on different cloud providers and systems.

