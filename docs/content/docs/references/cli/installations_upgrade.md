---
title: "porter installations upgrade"
slug: porter_installations_upgrade
url: /cli/porter_installations_upgrade/
---
## porter installations upgrade

Upgrade an installation

### Synopsis

Upgrade an installation.

The first argument is the installation name to upgrade. This defaults to the name of the bundle.

Porter uses the docker driver as the default runtime for executing a bundle image, but an alternate driver may be supplied via '--driver/-d' or the PORTER_RUNTIME_DRIVER environment variable.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.

The docker driver runs the bundle container using the local Docker host. To use a remote Docker host, set the following environment variables:
  DOCKER_HOST (required)
  DOCKER_TLS_VERIFY (optional)
  DOCKER_CERT_PATH (optional)


```
porter installations upgrade [INSTALLATION] [flags]
```

### Examples

```
  porter installation upgrade --version 0.2.0
  porter installation upgrade --reference ghcr.io/getporter/examples/kubernetes:v0.2.0
  porter installation upgrade --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter installation upgrade MyAppInDev --file myapp/bundle.json
  porter installation upgrade --parameter-set azure --param test-mode=true --param header-color=blue
  porter installation upgrade --param config=@config.json
  porter installation upgrade --credential-set azure --credential-set kubernetes
  porter installation upgrade --driver debug

```

### Options

```
      --allow-docker-host-access        Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://porter.sh/configuration/#allow-docker-host-access for the full implications of this flag.
      --autobuild-disabled              Do not automatically build the bundle from source when the last build is out-of-date.
      --cnab-file string                Path to the CNAB bundle.json file.
  -c, --credential-set stringArray      Credential sets to use when running the bundle. It should be a named set of credentials and may be specified multiple times.
      --debug                           Run the bundle in debug mode.
  -d, --driver string                   Specify a driver to use. Allowed values: docker, debug (default "docker")
  -f, --file porter.yaml                Path to the Porter manifest. Defaults to porter.yaml in the current directory.
      --force                           Force a fresh pull of the bundle
      --force-upgrade                   Force the upgrade to run even if the current installation is marked as failed.
  -h, --help                            help for upgrade
      --insecure-registry               Don't require TLS for the registry
      --mount-host-volume stringArray   Mount a host volume into the bundle. Format is <host path>:<container path>[:<option>]. May be specified multiple times. Option can be ro (read-only), rw (read-write), default is ro.
  -n, --namespace string                Namespace of the specified installation. Defaults to the global namespace.
      --no-logs                         Do not persist the bundle execution logs
      --param stringArray               Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times. For object parameters, use @FILEPATH to load JSON from a file (e.g., --param config=@config.json).
  -p, --parameter-set stringArray       Parameter sets to use when running the bundle. It should be a named set of parameters and may be specified multiple times.
  -r, --reference string                Use a bundle in an OCI registry specified by the given reference.
      --version string                  Version to which the installation should be upgraded. This represents the version of the bundle, which assumes the convention of setting the bundle tag to its version.
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter installations](/cli/porter_installations/)	 - Installation commands

