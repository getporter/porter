---
title: "porter installations show"
slug: porter_installations_show
url: /cli/porter_installations_show/
---
## porter installations show

Show an installation of a bundle

### Synopsis

Displays info relating to an installation of a bundle, including status and a listing of outputs.

```
porter installations show [INSTALLATION] [flags]
```

### Examples

```
  porter installation show
  porter installation show another-bundle

Optional output formats include json and yaml.

```

### Options

```
  -h, --help               help for show
  -n, --namespace string   Namespace in which the installation is defined. Defaults to the global namespace.
  -o, --output string      Specify an output format.  Allowed values: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter installations](/cli/porter_installations/)	 - Installation commands

