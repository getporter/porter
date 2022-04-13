---
title: "porter inspect"
slug: porter_inspect
url: /cli/porter_inspect/
---
## porter inspect

Inspect a bundle

### Synopsis

Inspect a bundle by printing the invocation images and any related images images.

If you would like more information about the bundle, the porter explain command will provide additional information,
like parameters, credentials, outputs and custom actions available.


```
porter inspect [flags]
```

### Examples

```
  porter inspect
  porter inspect --reference ghcr.io/getporter/examples/porter-hello:v0.2.0
  porter inspect --reference localhost:5000/ghcr.io/getporter/examples/porter-hello:v0.2.0 --insecure-registry --force
  porter inspect --file another/porter.yaml
  porter inspect --cnab-file some/bundle.json
		  
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
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter](/cli/porter/)	 - With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://porter.sh/quickstart to learn how to use Porter.


