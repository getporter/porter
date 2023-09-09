---
title: "porter logs"
slug: porter_logs
url: /cli/porter_logs/
---
## porter logs

Show the logs from an installation

### Synopsis

Show the logs from an installation.

Either display the logs from a specific run of a bundle with --run, or use --installation to display the logs from its most recent run.

```
porter logs [flags]
```

### Examples

```
  porter logs --installation wordpress --namespace dev
  porter installations logs show --run 01EZSWJXFATDE24XDHS5D5PWK6
```

### Options

```
  -h, --help                  help for logs
  -i, --installation string   The installation that generated the logs.
  -n, --namespace string      Namespace in which the installation is defined. Defaults to the global namespace.
  -r, --run string            The bundle run that generated the logs.
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


