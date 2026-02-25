---
title: "porter installations apply"
slug: porter_installations_apply
url: /cli/porter_installations_apply/
---
## porter installations apply

Apply changes to an installation

### Synopsis

Apply changes from the specified file to an installation. If the installation doesn't already exist, it is created.
The installation's bundle is automatically executed if changes are detected.

When the namespace is not set in the file, the current namespace is used.

You can use the show command to create the initial file:
  porter installation show mybuns --output yaml > mybuns.yaml


```
porter installations apply FILE [flags]
```

### Examples

```
  porter installation apply myapp.yaml
  porter installation apply myapp.yaml --dry-run
  porter installation apply myapp.yaml --force
```

### Options

```
      --dry-run            Evaluate if the bundle would be executed based on the changes in the file.
      --force              Force the bundle to be executed when no changes are detected.
  -h, --help               help for apply
  -n, --namespace string   Namespace in which the installation is defined. Defaults to the namespace defined in the file.
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter installations](/cli/porter_installations/)	 - Installation commands

