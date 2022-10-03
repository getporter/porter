---
title: "porter bundles inspect"
slug: porter_bundles_inspect
url: /cli/porter_bundles_inspect/
---
## porter bundles inspect

Inspect a bundle

### Synopsis

Inspect a bundle by printing the invocation images and any related images images.

If you would like more information about the bundle, the porter explain command will provide additional information,
like parameters, credentials, outputs and custom actions available.


```
porter bundles inspect REFERENCE [flags]
```

### Examples

```
  porter bundle inspect
  porter bundle inspect ghcr.io/getporter/examples/porter-hello:v0.2.0
  porter bundle inspect localhost:5000/ghcr.io/getporter/examples/porter-hello:v0.2.0 --insecure-registry --force
  porter bundle inspect --file another/porter.yaml
  porter bundle inspect --cnab-file some/bundle.json
		  
```

### Options

```
      --cnab-file string    Path to the CNAB bundle.json file.
  -f, --file porter.yaml    Path to the Porter manifest. Defaults to porter.yaml in the current directory.
      --force               Force a fresh pull of the bundle
  -h, --help                help for inspect
      --insecure-registry   Don't require TLS for the registry
  -o, --output string       Specify an output format.  Allowed values: plaintext, json, yaml (default "plaintext")
  -r, --reference string    Use a bundle in an OCI registry specified by the given reference.
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter bundles](/cli/porter_bundles/)	 - Bundle commands

