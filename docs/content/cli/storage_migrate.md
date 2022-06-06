---
title: "porter storage migrate"
slug: porter_storage_migrate
url: /cli/porter_storage_migrate/
---
## porter storage migrate

Migrate data from a previous version of Porter

### Synopsis

Migrate data from a previous version of Porter into your current installation of Porter.

Before running this command, you should have:

1. Installed the new version of Porter
2. Renamed the old PORTER_HOME directory, for example: mv ~/.porter ~/.porterv0
3. Created a new PORTER_HOME directory for the new version of Porter, for example: mkdir ~/.porter
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
  -n, --namespace string     Destination namespace where the migrated data should be saved.
      --old-account string   Name of the storage account in the old Porter configuration file containing the data that should be migrated. If unspecified, the default storage account is used.
      --old-home string      Path to the old PORTER_HOME directory where the previous version of Porter is installed
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter storage](/cli/porter_storage/)	 - Manage data stored by Porter

