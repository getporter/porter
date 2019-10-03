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
	porter publish --insecure
	porter publish --archive /tmp/mybuns.tgz
	porter publish --archive /tmp/mybuns.tgz --tag myrepo/my-buns:0.1.0
		
```

### Options

```
  -a, --archive string      Path to the bundle archive in .tgz format
  -f, --file porter.yaml    Path to the Porter manifest. Defaults to porter.yaml in the current directory.
  -h, --help                help for publish
      --insecure-registry   Don't require TLS for the registry.
  -t, --tag string          Bundle tag for newly published bundle
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

