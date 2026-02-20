---
title: Multiple Configuration Environments
description: Manage multiple Porter environments in a single config file using named contexts
weight: 5
---

Porter supports multiple named configuration contexts in a single config file,
similar to how `kubectl` handles multiple clusters.
This lets you switch between environments — such as `dev`, `staging`, and `prod`
— without maintaining separate config files.

- [Config file format](#config-file-format)
- [Selecting a context](#selecting-a-context)
- [Managing contexts](#managing-contexts)
  - [List contexts](#list-contexts)
  - [Switch context](#switch-context)
- [Migrating from the legacy format](#migrating-from-the-legacy-format)

## Config file format

The multi-context format requires `schemaVersion: "2.0.0"` at the top of the
config file. All settings are placed under named contexts inside a `contexts`
list, instead of at the top level.

```yaml
# ~/.porter/config.yaml
schemaVersion: "2.0.0"

# The context used when --context is not specified
current-context: default

contexts:
  - name: default
    config:
      namespace: dev
      default-storage: devdb
      default-secrets-plugin: filesystem
      storage:
        - name: devdb
          plugin: mongodb
          config:
            url: "mongodb://localhost:27017/porter"

  - name: prod
    config:
      namespace: prod
      default-storage: proddb
      default-secrets-plugin: azure.keyvault
      storage:
        - name: proddb
          plugin: mongodb
          config:
            url: "${secret.prod-db-connection-string}"
      secrets:
        - name: prodvault
          plugin: azure.keyvault
          config:
            vault: "my-prod-vault"
            subscription-id: "${env.AZURE_SUBSCRIPTION_ID}"
```

The `config` block inside each context entry accepts exactly the same settings
as the legacy flat config format.
See [Configuration](/docs/configuration/configuration/) for a full list of
available settings.

## Selecting a context

Porter resolves the active context using the following priority order
(highest to lowest):

| Source | How to set |
|---|---|
| `--context` flag | `porter install --context prod` |
| `PORTER_CONTEXT` environment variable | `export PORTER_CONTEXT=prod` |
| `current-context` field in the config file | `porter config context use prod` |
| Falls back to a context named `default` | (automatic) |

```bash
# Use the prod context for a single command
porter install --context prod

# Use the prod context for the rest of the shell session
export PORTER_CONTEXT=prod
porter install
porter list
```

## Managing contexts

### List contexts

`porter config context list` prints all contexts defined in the config file.
The currently active context is marked with `*`.

```console
$ porter config context list
* default
  prod
  staging
```

The active context reflects the same priority order described above.
Passing `--context` changes which entry is marked:

```console
$ porter config context list --context prod
  default
* prod
  staging
```

### Switch context

`porter config context use <name>` updates the `current-context` field in the
config file so that subsequent commands use the specified context by default.

```console
$ porter config context use prod
Switched to context "prod".

$ porter config context list
  default
* prod
  staging
```

## Migrating from the legacy format

If you have an existing flat config file, run `porter config migrate` to
convert it automatically.
The command wraps your current settings in a context named `default` and adds
the required `schemaVersion` header.

```console
$ porter config migrate
Migrated config.yaml to multi-context format. Use 'porter config show' to review.
```

The migration is text-based, so template variables such as
`${env.MY_VAR}` and `${secret.MY_SECRET}` are preserved exactly as written.

Only YAML config files are supported for automatic migration.
For TOML, JSON, or HCL files, `porter config migrate` prints the required
structure so you can apply the changes manually.

After migrating, verify the result with `porter config show` and add additional
contexts as needed.
