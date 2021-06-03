---
title: "porter storage migrate"
slug: porter_storage_migrate
url: /cli/porter_storage_migrate/
---
## porter storage migrate

Migrate active storage account

### Synopsis

Migrate the data in the active storage account to the schema used by this version of Porter.

Always back up Porter's data before performing a migration. Instructions for backing up are at https://porter.sh/storage-migrate.

```
porter storage migrate [flags]
```

### Options

```
  -h, --help   help for migrate
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter storage](/cli/porter_storage/)	 - Manage data stored by Porter

