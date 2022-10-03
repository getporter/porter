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
  porter copy --source ghcr.io/getporter/examples/porter-hello:v0.2.0 --destination portersh
  porter copy --source ghcr.io/getporter/examples/porter-hello:v0.2.0 --destination portersh --insecure-registry
		  
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

* [porter](/cli/porter/)	 - With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://getporter.org/quickstart to learn how to use Porter.


