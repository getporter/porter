---
title: "porter installations delete"
slug: porter_installations_delete
url: /cli/porter_installations_delete/
---
## porter installations delete

Delete an installation

### Synopsis

Deletes all records and outputs associated with an installation

```
porter installations delete [INSTALLATION] [flags]
```

### Examples

```
  porter installation delete
porter installation delete wordpress
porter installation delete --force

```

### Options

```
      --force   Force a delete the installation, regardless of last completed action
  -h, --help    help for delete
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter installations](/cli/porter_installations/)	 - Installation commands

