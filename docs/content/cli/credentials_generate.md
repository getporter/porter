---
title: "porter credentials generate"
slug: porter_credentials_generate
url: /cli/porter_credentials_generate/
---
## porter credentials generate

Generate Credential Set

### Synopsis

Generate a named set of credentials.

The first argument is the name of credential set you wish to generate. If not
provided, this will default to the bundle name. By default, Porter will
generate a credential set for the bundle in the current directory. You may also
specify a bundle with --file.

Bundles define 1 or more credential(s) that are required to interact with a
bundle. The bundle definition defines where the credential should be delivered
to the bundle, i.e. at /root/.kube. A credential set, on the other hand,
represents the source data that you wish to use when interacting with the
bundle. These will typically be environment variables or files on your local
file system.

When you wish to install, upgrade or delete a bundle, Porter will use the
credential set to determine where to read the necessary information from and
will then provide it to the bundle in the correct location. 

```
porter credentials generate [NAME] [flags]
```

### Examples

```
  porter credential generate
  porter credential generate kubecred --tag getporter/porter-hello:v0.1.0
  porter credential generate kubecred --tag localhost:5000/getporter/porter-hello:v0.1.0 --insecure-registry --force
  porter credential generate kubecred --file myapp/porter.yaml
  porter credential generate kubecred --cnab-file myapp/bundle.json

```

### Options

```
      --cnab-file string    Path to the CNAB bundle.json file.
  -f, --file string         Path to the porter manifest file. Defaults to the bundle in the current directory.
      --force               Force a fresh pull of the bundle
  -h, --help                help for generate
      --insecure-registry   Don't require TLS for the registry
      --tag string          Use a bundle in an OCI registry specified by the given tag.
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter credentials](/cli/porter_credentials/)	 - Credentials commands

