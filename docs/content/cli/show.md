---
title: "porter show"
slug: porter_show
url: /cli/porter_show/
---
## porter show

Show an installation of a bundle

### Synopsis

Displays info relating to an installation of a bundle, including status and a listing of outputs.

```
porter show [INSTALLATION] [flags]
```

### Examples

```
  porter show
  porter show another-bundle

Optional output formats include json and yaml.

```

### Options

```
  -h, --help               help for show
  -n, --namespace string   Namespace in which the installation is defined. Defaults to the global namespace.
  -o, --output string      Specify an output format.  Allowed values: plaintext, json, yaml (default "plaintext")
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter](/cli/porter/)	 - With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://getporter.org/quickstart to learn how to use Porter.


