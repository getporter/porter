---
title: "porter copy"
slug: porter_copy
url: /cli/porter_copy/
---
## porter copy

Copy a bundle

### Synopsis

Copy a published bundle from one registry to another.		
Source bundle can be either a tagged reference or a digest reference.
Destination can be either a registry, a registry/repository, or a fully tagged bundle reference. 
If the source bundle is a digest reference, destination must be a tagged reference.


```
porter copy [flags]
```

### Examples

```
  porter copy
  porter copy --source deislabs/porter-bundle:v0.1.0 --destination portersh
  porter copy --source deislabs/porter-bundle:v0.1.0 --destination portersh --insecure-registry
		  
```

### Options

```
      --destination string   The registry to copy the bundle to. Can be registry name, registry plus a repo prefix, or a new tagged reference. All images and the bundle will be prefixed with registry.
  -h, --help                 help for copy
      --insecure-registry    Don't require TLS for registries
      --source string         The fully qualified source bundle, including tag or digest.
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

