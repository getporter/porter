---
title: "porter bundles explain"
slug: porter_bundles_explain
url: /cli/porter_bundles_explain/
---
## porter bundles explain

Explain a bundle

### Synopsis

Explain how to use a bundle by printing the parameters, credentials, outputs, actions.

```
porter bundles explain REFERENCE [flags]
```

### Examples

```
  porter bundle explain
  porter bundle explain ghcr.io/getporter/examples/porter-hello:v0.2.0
  porter bundle explain localhost:5000/ghcr.io/getporter/examples/porter-hello:v0.2.0 --insecure-registry --force
  porter bundle explain --file another/porter.yaml
  porter bundle explain --cnab-file some/bundle.json
  porter bundle explain --action install
		  
```

### Options

```
      --action string       Hide parameters and outputs that are not used by the specified action.
      --cnab-file string    Path to the CNAB bundle.json file.
  -f, --file porter.yaml    Path to the Porter manifest. Defaults to porter.yaml in the current directory.
      --force               Force a fresh pull of the bundle
  -h, --help                help for explain
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

