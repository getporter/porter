---
title: "porter explain"
slug: porter_explain
url: /cli/porter_explain/
---
## porter explain

Explain a bundle

### Synopsis

Explain how to use a bundle by printing the parameters, credentials, outputs, actions.

```
porter explain [flags]
```

### Examples

```
  porter explain
  porter explain --reference ghcr.io/getporter/examples/porter-hello:v0.2.0
  porter explain --reference localhost:5000/ghcr.io/getporter/examples/porter-hello:v0.2.0 --insecure-registry --force
  porter explain --file another/porter.yaml
  porter explain --cnab-file some/bundle.json
  porter explain --action install
		  
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
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter](/cli/porter/)	 - With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://getporter.org/quickstart to learn how to use Porter.


