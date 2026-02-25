---
title: "porter mixins uninstall"
slug: porter_mixins_uninstall
url: /cli/porter_mixins_uninstall/
---
## porter mixins uninstall

Uninstall a mixin

```
porter mixins uninstall NAME [flags]
```

### Examples

```
  porter mixin uninstall helm
```

### Options

```
  -h, --help   help for uninstall
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter mixins](/cli/porter_mixins/)	 - Mixin commands. Mixins assist with authoring bundles.

