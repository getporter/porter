---
title: "porter bundles invoke"
slug: porter_bundles_invoke
url: /cli/porter_bundles_invoke/
---
## porter bundles invoke

Invoke a custom action on an installation

### Synopsis

Invoke a custom action on an installation.

The first argument is the installation name upon which to invoke the action. This defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.

```
porter bundles invoke [INSTALLATION] --action ACTION [flags]
```

### Examples

```
  porter bundle invoke --action ACTION
  porter bundle invoke --reference getporter/kubernetes:v0.1.0
  porter bundle invoke --reference localhost:5000/getporter/kubernetes:v0.1.0 --insecure-registry --force
  porter bundle invoke --action ACTION MyAppInDev --file myapp/bundle.json
  porter bundle invoke --action ACTION  --parameter-set azure --param test-mode=true --param header-color=blue
  porter bundle invoke --action ACTION --cred azure --cred kubernetes
  porter bundle invoke --action ACTION --driver debug

```

### Options

```
      --action string              Custom action name to invoke.
      --allow-docker-host-access   Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://porter.sh/configuration/#allow-docker-host-access for the full implications of this flag.
      --cnab-file string           Path to the CNAB bundle.json file.
  -c, --cred strings               Credential to use when installing the bundle. May be either a named set of credentials or a filepath, and specified multiple times.
  -d, --driver string              Specify a driver to use. Allowed values: docker, debug (default "docker")
  -f, --file string                Path to the porter manifest file. Defaults to the bundle in the current directory.
      --force                      Force a fresh pull of the bundle
  -h, --help                       help for invoke
      --insecure-registry          Don't require TLS for the registry
  -n, --namespace string           Namespace of the specified installation. Defaults to the global namespace.
      --no-logs                    Do not persist the bundle execution logs
      --param strings              Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times.
  -p, --parameter-set strings      Name of a parameter set file for the bundle. May be either a named set of parameters or a filepath, and specified multiple times.
  -r, --reference string           Use a bundle in an OCI registry specified by the given reference.
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter bundles](/cli/porter_bundles/)	 - Bundle commands

