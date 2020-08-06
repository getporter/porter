---
title: "porter bundles archive"
slug: porter_bundles_archive
url: /cli/porter_bundles_archive/
---
## porter bundles archive

Archive a bundle from a tag

### Synopsis

Archives a bundle by generating a gzipped tar archive containing the bundle, invocation image and any referenced images.

```
porter bundles archive FILENAME --tag PUBLISHED_BUNDLE [flags]
```

### Examples

```
  porter bundle archive mybun.tgz --tag getporter/porter-hello:v0.1.0
  porter bundle archive mybun.tgz --tag localhost:5000/getporter/porter-hello:v0.1.0 --force

```

### Options

```
      --force               Force a fresh pull of the bundle
  -h, --help                help for archive
      --insecure-registry   Don't require TLS for the registry
      --tag string          Use a bundle in an OCI registry specified by the given tag.
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter bundles](/cli/porter_bundles/)	 - Bundle commands

