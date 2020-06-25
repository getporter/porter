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
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter installations output](/cli/porter_installations_output/)	 - Output commands

