---
title: "porter installations delete"
slug: porter_installations_delete
url: /cli/porter_installations_delete/
---
## porter installations delete

Delete an installation

### Synopsis

Deletes all records and outputs associated with an installation

```
porter installations delete [INSTALLATION] [flags]
```

### Examples

```
  porter installation delete
  porter installation delete wordpress
  porter installation delete --force

```

### Options

```
      --force              Force a delete the installation, regardless of last completed action
  -h, --help               help for delete
  -n, --namespace string   Namespace in which the installation is defined. Defaults to the global namespace.
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter installations](/cli/porter_installations/)	 - Installation commands

