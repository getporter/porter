---
title: "porter installations output show"
slug: porter_installations_output_show
url: /cli/porter_installations_output_show/
---
## porter installations output show

Show the output of an installation

### Synopsis

Show the output of an installation

```
porter installations output show NAME [--installation|-i INSTALLATION] [flags]
```

### Examples

```
  porter installation output show kubeconfig
    porter installation output show subscription-id --installation azure-mysql
```

### Options

```
  -h, --help                  help for show
  -i, --installation string   Specify the installation to which the output belongs.
  -n, --namespace string      Namespace in which the installation is defined. Defaults to the global namespace.
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter installations output](/cli/porter_installations_output/)	 - Output commands

