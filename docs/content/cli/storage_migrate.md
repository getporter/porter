---
title: "porter storage migrate"
slug: porter_storage_migrate
url: /cli/porter_storage_migrate/
---
## porter storage migrate

Migrate data from v0.38 to v1

### Synopsis

Migrate data from Porter v0.38 into a v1 installation of Porter.

See https://getporter.org/storage-migrate for a full description of the migration process. Below is a summary:

Before running this command, you should have:

1. Installed Porter v1, see https://getporter.org/install for instructions.
2. Renamed the old PORTER_HOME directory, for example: mv ~/.porter ~/.porterv0.
3. Created a new PORTER_HOME directory for the new version of Porter, for example: mkdir ~/.porter.
4. Configured a default storage account for the new version of Porter. The data from the old account will be migrated into the default storage account.

This upgrades the data to the current storage schema, and does not change the data stored in the old account.

```
porter storage migrate --old-home OLD_PORTER_HOME [--old-account STORAGE_NAME] [--namespace NAMESPACE] [flags]
```

### Examples

```
  porter storage migrate --old-home ~/.porterv0
  porter storage migrate --old-account my-azure --old-home ~/.porterv0
  porter storage migrate --namespace new-namespace --old-home ~/.porterv0

```

### Options

```
  -h, --help                 help for migrate
  -n, --namespace string     Destination namespace where the migrated data should be saved. By default, Porter migrates your data into the current namespace as configured by environment variables and your config file, otherwise the global namespace is used.
      --old-account string   Name of the storage account in the old Porter configuration file containing the data that should be migrated. If unspecified, the default storage account is used.
      --old-home string      Path to the old PORTER_HOME directory where the previous version of Porter is installed
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter storage](/cli/porter_storage/)	 - Manage data stored by Porter

