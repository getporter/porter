---
title: "porter config migrate"
slug: porter_config_migrate
url: /cli/porter_config_migrate/
---
## porter config migrate

Migrate the config file to the multi-context format

### Synopsis

Migrate the porter config file from the legacy flat format to the
multi-context format (schemaVersion: "2.0.0"). The existing settings are
preserved under a context named "default".

Only YAML config files are supported for automatic migration. For TOML,
JSON, or HCL files, the required structure is printed so you can apply
the changes manually.

```
porter config migrate [flags]
```

### Examples

```
  porter config migrate
```

### Options

```
  -h, --help   help for migrate
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter config](/cli/porter_config/)	 - Config commands

