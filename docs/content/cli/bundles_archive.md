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
porter bundles archive [flags]
```

### Examples

```
  porter bundle archive [FILENAME]
  porter bundle archive --file another/porter.yaml [FILENAME]
  porter bundle archive --cnab-file some/bundle.json [FILENAME]
  porter bundle archive --tag repo/bundle:tag [FILENAME]
		  
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

