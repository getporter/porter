---
title: "porter bundles copy"
slug: porter_bundles_copy
url: /cli/porter_bundles_copy/
---
## porter bundles copy

Copy a bundle

### Synopsis

Copy a published bundle from one registry to another.		
Source bundle can be either a tagged reference or a digest reference.
Destination can be either a registry, a registry/repository, or a fully tagged bundle reference. 
If the source bundle is a digest reference, destination must be a tagged reference.


```
porter bundles copy [flags]
```

### Examples

```
  porter bundle copy
  porter bundle copy --source ghcr.io/getporter/examples/porter-hello:v0.2.0 --destination portersh
  porter bundle copy --source ghcr.io/getporter/examples/porter-hello:v0.2.0 --destination portersh --insecure-registry
		  
```

### Options

```
      --destination string   The registry to copy the bundle to. Can be registry name, registry plus a repo prefix, or a new tagged reference. All images and the bundle will be prefixed with registry.
      --force                Force push the bundle to overwrite the previously published bundle
  -h, --help                 help for copy
      --insecure-registry    Don't require TLS for registries
      --source string         The fully qualified source bundle, including tag or digest.
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter bundles](/cli/porter_bundles/)	 - Bundle commands

