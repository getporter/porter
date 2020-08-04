---
title: "porter archive"
slug: porter_archive
url: /cli/porter_archive/
---
## porter archive

Archive a bundle from a tag

### Synopsis

Archives a bundle by generating a gzipped tar archive containing the bundle, invocation image and any referenced images.

```
porter archive FILENAME --tag PUBLISHED_BUNDLE [flags]
```

### Examples

```
  porter archive mybun.tgz --tag repo/bundle:tag
```

### Options

```
      --force        Force a fresh pull of the bundle
  -h, --help         help for archive
  -t, --tag string   Use a bundle in an OCI registry specified by the given tag
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

