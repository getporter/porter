---
title: "porter bundles upgrade"
slug: porter_bundles_upgrade
url: /cli/porter_bundles_upgrade/
---
## porter bundles upgrade

Upgrade a bundle instance

### Synopsis

Upgrade a bundle instance.

The first argument is the bundle instance name to upgrade. This defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.

```
porter bundles upgrade [INSTANCE] [flags]
```

### Examples

```
  porter bundle upgrade
  porter bundle upgrade --insecure
  porter bundle upgrade MyAppInDev --file myapp/bundle.json
  porter bundle upgrade --param-file base-values.txt --param-file dev-values.txt --param test-mode=true --param header-color=blue
  porter bundle upgrade --cred azure --cred kubernetes
  porter bundle upgrade --driver debug
  porter bundle upgrade MyAppFromTag --tag deislabs/porter-kube-bundle:v1.0

```

### Options

```
      --cnab-file string     Path to the CNAB bundle.json file.
  -c, --cred strings         Credential to use when installing the bundle. May be either a named set of credentials or a filepath, and specified multiple times.
  -d, --driver string        Specify a driver to use. Allowed values: docker, debug (default "docker")
  -f, --file string          Path to the porter manifest file. Defaults to the bundle in the current directory.
  -h, --help                 help for upgrade
      --insecure             Allow working with untrusted bundles (default true)
      --insecure-registry    Don't require TLS for the registry
      --param strings        Define an individual parameter in the form NAME=VALUE. Overrides parameters set with the same name using --param-file. May be specified multiple times.
      --param-file strings   Path to a parameters definition file for the bundle, each line in the form of NAME=VALUE. May be specified multiple times.
  -t, --tag string           Use a bundle in an OCI registry specified by the given tag
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter bundles](/cli/porter_bundles/)	 - Bundle commands

