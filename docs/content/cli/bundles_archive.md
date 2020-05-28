---
title: "porter bundles archive"
slug: porter_bundles_archive
url: /cli/porter_bundles_archive/
---
## porter bundles archive

Archive a bundle

### Synopsis

Archives a bundle by generating a gzipped tar archive containing the bundle, invocation image and any referenced images.

```
porter bundles archive FILENAME [flags]
```

### Examples

```
  porter bundle archive mybun.tgz
  porter bundle archive mybun.tgz --file another/porter.yaml
  porter bundle archive mybun.tgz --cnab-file some/bundle.json
  porter bundle archive mybun.tgz --tag repo/bundle:tag
		  
```

### Options

```
      --cnab-file string   Path to the CNAB bundle.json file.
  -f, --file porter.yaml   Path to the Porter manifest. Defaults to porter.yaml in the current directory.
      --force              Force a fresh pull of the bundle
  -h, --help               help for archive
  -t, --tag string         Use a bundle in an OCI registry specified by the given tag
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter bundles](/cli/porter_bundles/)	 - Bundle commands

