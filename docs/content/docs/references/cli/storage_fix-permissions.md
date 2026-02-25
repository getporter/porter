---
title: "porter storage fix-permissions"
slug: porter_storage_fix-permissions
url: /cli/porter_storage_fix-permissions/
---
## porter storage fix-permissions

Fix the permissions on your PORTER_HOME directory

### Synopsis

This will reset the permissions on your PORTER_HOME directory to the least permissions required, where only the current user has permissions.

```
porter storage fix-permissions [flags]
```

### Options

```
  -h, --help   help for fix-permissions
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter storage](/cli/porter_storage/)	 - Manage data stored by Porter

