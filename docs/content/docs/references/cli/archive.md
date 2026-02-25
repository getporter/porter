---
title: "porter archive"
slug: porter_archive
url: /cli/porter_archive/
---
## porter archive

Archive a bundle from a reference

### Synopsis

Archives a bundle by generating a gzipped tar archive containing the bundle, bundle image and any referenced images.

```
porter archive FILENAME --reference PUBLISHED_BUNDLE [flags]
```

### Examples

```
  porter archive mybun.tgz --reference ghcr.io/getporter/examples/porter-hello:v0.2.0
  porter archive mybun.tgz --reference localhost:5000/ghcr.io/getporter/examples/porter-hello:v0.2.0 --force
  porter archive mybun.tgz --compression NoCompression --reference ghcr.io/getporter/examples/porter-hello:v0.2.0

```

### Options

```
  -c, --compression string   Compression level to use when creating the gzipped tar archive. Allowed values are: BestCompression, BestSpeed, DefaultCompression, HuffmanOnly, NoCompression (default "DefaultCompression")
      --force                Force a fresh pull of the bundle
  -h, --help                 help for archive
      --insecure-registry    Don't require TLS for the registry
  -r, --reference string     Use a bundle in an OCI registry specified by the given reference.
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter](/cli/porter/)	 - With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://porter.sh/quickstart to learn how to use Porter.


