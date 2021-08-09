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
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

