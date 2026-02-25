---
title: "porter plugins search"
slug: porter_plugins_search
url: /cli/porter_plugins_search/
---
## porter plugins search

Search available plugins

### Synopsis

Search available plugins. You can specify an optional plugin name query, where the results are filtered by plugins whose name contains the query term.

By default the community plugin index at https://cdn.porter.sh/plugins/index.json is searched. To search from a mirror, set the environment variable PORTER_MIRROR, or mirror in the Porter config file, with the value to replace https://cdn.porter.sh with.

```
porter plugins search [QUERY] [flags]
```

### Examples

```
  porter plugin search
  porter plugin search azure
  porter plugin search -o json
```

### Options

```
  -h, --help            help for search
      --mirror string   Mirror of official Porter assets (default "https://cdn.porter.sh")
  -o, --output string   Output format, allowed values are: plaintext, json, yaml (default "plaintext")
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter plugins](/cli/porter_plugins/)	 - Plugin commands. Plugins enable Porter to work on different cloud providers and systems.

