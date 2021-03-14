---
title: "porter installations logs show"
slug: porter_installations_logs_show
url: /cli/porter_installations_logs_show/
---
## porter installations logs show

Show the logs from an installation

### Synopsis

Show the logs from an installation.

Either display the logs from a specific run of a bundle with --run, or use --installation to display the logs from its most recent run.

```
porter installations logs show [flags]
```

### Examples

```
  porter installation logs show --installation wordpress
  porter installations logs show --run 01EZSWJXFATDE24XDHS5D5PWK6
```

### Options

```
  -h, --help                  help for show
  -i, --installation string   The installation that generated the logs.
  -r, --run string            The bundle run that generated the logs.
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter installations logs](/cli/porter_installations_logs/)	 - Installation Logs commands

