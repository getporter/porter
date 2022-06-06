---
title: "porter storage migrate"
slug: porter_storage_migrate
url: /cli/porter_storage_migrate/
---
## porter storage migrate

Migrate data from an older version of Porter

### Synopsis

Copies data from a source storage account defined in Porter's config file into a destination storage account. 

This upgrades the data to the current storage schema, and does not change the data stored in the source account.

```
porter storage migrate --src OLD_ACCOUNT --dest NEW_ACCOUNT [flags]
```

### Options

```
  -d, --dest string   Name of the destination storage account defined in your Porter config file
  -h, --help          help for migrate
  -s, --src string    Name of the source storage account defined in your Porter config file
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter storage](/cli/porter_storage/)	 - Manage data stored by Porter

