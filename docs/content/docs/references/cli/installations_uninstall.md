---
title: "porter installations uninstall"
slug: porter_installations_uninstall
url: /cli/porter_installations_uninstall/
---
## porter installations uninstall

Uninstall an installation

### Synopsis

Uninstall an installation

The first argument is the installation name to uninstall. This defaults to the name of the bundle.

Porter uses the docker driver as the default runtime for executing a bundle image, but an alternate driver may be supplied via '--driver/-d'' or the PORTER_RUNTIME_DRIVER environment variable.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.

The docker driver runs the bundle container using the local Docker host. To use a remote Docker host, set the following environment variables:
  DOCKER_HOST (required)
  DOCKER_TLS_VERIFY (optional)
  DOCKER_CERT_PATH (optional)


```
porter installations uninstall [INSTALLATION] [flags]
```

### Examples

```
  porter installation uninstall
  porter installation uninstall --reference ghcr.io/getporter/examples/kubernetes:v0.2.0
  porter installation uninstall --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter installation uninstall MyAppInDev --file myapp/bundle.json
  porter installation uninstall --parameter-set azure --param test-mode=true --param header-color=blue
  porter installation uninstall --credential-set azure --credential-set kubernetes
  porter installation uninstall --driver debug
  porter installation uninstall --delete
  porter installation uninstall --force-delete

```

### Options

```
      --allow-docker-host-access        Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://porter.sh/configuration/#allow-docker-host-access for the full implications of this flag.
      --autobuild-disabled              Do not automatically build the bundle from source when the last build is out-of-date.
      --cnab-file string                Path to the CNAB bundle.json file.
  -c, --credential-set stringArray      Credential sets to use when running the bundle. It should be a named set of credentials and may be specified multiple times.
      --debug                           Run the bundle in debug mode.
      --delete                          Delete all records associated with the installation, assuming the uninstall action succeeds
  -d, --driver string                   Specify a driver to use. Allowed values: docker, debug (default "docker")
  -f, --file porter.yaml                Path to the Porter manifest. Defaults to porter.yaml in the current directory.
      --force                           Force a fresh pull of the bundle
      --force-delete                    UNSAFE. Delete all records associated with the installation, even if uninstall fails. This is intended for cleaning up test data and is not recommended for production environments.
  -h, --help                            help for uninstall
      --insecure-registry               Don't require TLS for the registry
      --mount-host-volume stringArray   Mount a host volume into the bundle. Format is <host path>:<container path>[:<option>]. May be specified multiple times. Option can be ro (read-only), rw (read-write), default is ro.
  -n, --namespace string                Namespace of the specified installation. Defaults to the global namespace.
      --no-logs                         Do not persist the bundle execution logs
      --param stringArray               Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times. For object parameters, use @FILEPATH to load JSON from a file (e.g., --param config=@config.json).
  -p, --parameter-set stringArray       Parameter sets to use when running the bundle. It should be a named set of parameters and may be specified multiple times.
  -r, --reference string                Use a bundle in an OCI registry specified by the given reference.
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter installations](/cli/porter_installations/)	 - Installation commands

