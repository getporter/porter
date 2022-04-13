---
title: "porter bundles upgrade"
slug: porter_bundles_upgrade
url: /cli/porter_bundles_upgrade/
---
## porter bundles upgrade

Upgrade an installation

### Synopsis

Upgrade an installation.

The first argument is the installation name to upgrade. This defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d' or the PORTER_RUNTIME_DRIVER environment variable.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.

```
porter bundles upgrade [INSTALLATION] [flags]
```

### Examples

```
  porter bundle upgrade --version 0.2.0
  porter bundle upgrade --reference ghcr.io/getporter/examples/kubernetes:v0.2.0
  porter bundle upgrade --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter bundle upgrade MyAppInDev --file myapp/bundle.json
  porter bundle upgrade --parameter-set azure --param test-mode=true --param header-color=blue
  porter bundle upgrade --cred azure --cred kubernetes
  porter bundle upgrade --driver debug

```

### Options

```
      --allow-docker-host-access   Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://porter.sh/configuration/#allow-docker-host-access for the full implications of this flag.
      --cnab-file string           Path to the CNAB bundle.json file.
  -c, --cred strings               Credential to use when installing the bundle. It should be a named set of credentials and may be specified multiple times.
  -d, --driver string              Specify a driver to use. Allowed values: docker, debug (default "docker")
  -f, --file string                Path to the porter manifest file. Defaults to the bundle in the current directory.
      --force                      Force a fresh pull of the bundle
  -h, --help                       help for upgrade
      --insecure-registry          Don't require TLS for the registry
  -n, --namespace string           Namespace of the specified installation. Defaults to the global namespace.
      --no-logs                    Do not persist the bundle execution logs
      --param strings              Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times.
  -p, --parameter-set strings      Name of a parameter set file for the bundle. May be either a named set of parameters or a filepath, and specified multiple times.
  -r, --reference string           Use a bundle in an OCI registry specified by the given reference.
      --version string             Version to which the installation should be upgraded. This represents the version of the bundle, which assumes the convention of setting the bundle tag to its version.
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter bundles](/cli/porter_bundles/)	 - Bundle commands

