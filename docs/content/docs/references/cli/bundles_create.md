---
title: "porter bundles create"
slug: porter_bundles_create
url: /cli/porter_bundles_create/
---
## porter bundles create

Create a bundle

### Synopsis

Create a bundle. This command creates a new porter bundle with the specified bundle-name, in the directory with the specified bundle-name. The directory will be created if it doesn't already exist. If no bundle-name is provided, the bundle will be created in current directory and the bundle name will be 'porter-hello'.

```
porter bundles create [bundle-name] [flags]
```

### Options

```
  -h, --help   help for create
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter bundles](/cli/porter_bundles/)	 - Bundle commands

