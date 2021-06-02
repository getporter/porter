---
title: "porter publish"
slug: porter_publish
url: /cli/porter_publish/
---
## porter publish

Publish a bundle

### Synopsis

Publishes a bundle by pushing the invocation image and bundle to a registry.

Note: if overrides for registry/tag/reference are provided, this command only re-tags the invocation image and bundle; it does not re-build the bundle.

```
porter publish [flags]
```

### Examples

```
  porter publish
  porter publish --file myapp/porter.yaml
  porter publish --dir myapp
  porter publish --archive /tmp/mybuns.tgz --reference myrepo/my-buns:0.1.0
  porter publish --tag latest
  porter publish --registry myregistry.com/myorg
		
```

### Options

```
  -a, --archive string      Path to the bundle archive in .tgz format
  -d, --dir string          Path to the build context directory where all bundle assets are located.
  -f, --file porter.yaml    Path to the Porter manifest. Defaults to porter.yaml in the current directory.
  -h, --help                help for publish
      --insecure-registry   Don't require TLS for the registry
  -r, --reference string    Use a bundle in an OCI registry specified by the given reference.
      --registry string     Override the registry portion of the bundle reference, e.g. docker.io, myregistry.com/myorg
      --tag string          Override the Docker tag portion of the bundle reference, e.g. latest, v0.1.1
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

