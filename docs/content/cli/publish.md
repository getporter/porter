---
title: "porter publish"
slug: porter_publish
url: /cli/porter_publish/
---
## porter publish

Publish a bundle

### Synopsis

Publishes a bundle by pushing the invocation image and bundle to a registry.

```
porter publish [flags]
```

### Examples

```
  porter publish
  porter publish --file myapp/porter.yaml
  porter publish --archive /tmp/mybuns.tgz --reference myrepo/my-buns:0.1.0
		
```

### Options

```
  -a, --archive string      Path to the bundle archive in .tgz format
  -f, --file porter.yaml    Path to the Porter manifest. Defaults to porter.yaml in the current directory.
  -h, --help                help for publish
      --insecure-registry   Don't require TLS for the registry
      --reference string    Use a bundle in an OCI registry specified by the given reference.
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

