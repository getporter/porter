---
title: "porter installations output show"
slug: porter_installations_output_show
url: /cli/porter_installations_output_show/
---
## porter installations output show

Show the output of an installation

### Synopsis

Show the output of an installation.

Either display the output from a specific run of a bundle with --run, or use --installation to display the output from its most recent run.

```
porter installations output show NAME [--installation|-i INSTALLATION] [flags]
```

### Examples

```
  porter installation output show kubeconfig
    porter installation output show subscription-id --installation azure-mysql
    porter installation output show kubeconfig --run 01EZSWJXFATDE24XDHS5D5PWK6
```

### Options

```
  -h, --help                  help for show
  -i, --installation string   Specify the installation to which the output belongs.
  -n, --namespace string      Namespace in which the installation is defined. Defaults to the global namespace.
  -r, --run string            The bundle run that generated the output.
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter installations output](/cli/porter_installations_output/)	 - Output commands

