---
title: "porter installations apply"
slug: porter_installations_apply
url: /cli/porter_installations_apply/
---
## porter installations apply

Apply changes to an installation

### Synopsis

Apply changes from the specified file to an installation. If the installation doesn't already exist, it is created.

When the namespace is not set in the file, the current namespace is used.

You can use the show command to create the initial file:
  porter installation show mybuns --output yaml > mybuns.yaml


```
porter installations apply FILE [flags]
```

### Examples

```
  porter installation apply myapp.yaml
```

### Options

```
  -h, --help               help for apply
  -n, --namespace string   Namespace in which the installation is defined. Defaults to the namespace defined in the file.
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter installations](/cli/porter_installations/)	 - Installation commands

