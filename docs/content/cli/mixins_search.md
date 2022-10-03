---
title: "porter mixins search"
slug: porter_mixins_search
url: /cli/porter_mixins_search/
---
## porter mixins search

Search available mixins

### Synopsis

Search available mixins. You can specify an optional mixin name query, where the results are filtered by mixins whose name contains the query term.

By default the community mixin index at https://cdn.porter.sh/mixins/index.json is searched. To search from a mirror, set the environment variable PORTER_MIRROR, or mirror in the Porter config file, with the value to replace https://cdn.porter.sh with.

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
      --mirror string   Mirror of official Porter assets (default "https://cdn.porter.sh")
  -o, --output string   Output format, allowed values are: plaintext, json, yaml (default "plaintext")
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter mixins](/cli/porter_mixins/)	 - Mixin commands. Mixins assist with authoring bundles.

