---
title: "porter installations install"
slug: porter_installations_install
url: /cli/porter_installations_install/
---
## porter installations install

Create a new installation of a bundle

### Synopsis

Create a new installation of a bundle.

The first argument is the name of the installation to create. This defaults to the name of the bundle. 

Once a bundle has been successfully installed, the install action cannot be repeated. This is a precaution to avoid accidentally overwriting an existing installation. If you need to re-run install, which is common when authoring a bundle, you can use the --force flag to by-pass this check.

Porter uses the docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d' or the PORTER_RUNTIME_DRIVER environment variable.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.

The docker driver runs the bundle container using the local Docker host. To use a remote Docker host, set the following environment variables:
  DOCKER_HOST (required)
  DOCKER_TLS_VERIFY (optional)
  DOCKER_CERT_PATH (optional)


```
porter installations install [INSTALLATION] [flags]
```

### Examples

```
  porter installation install
  porter installation install MyAppFromReference --reference ghcr.io/getporter/examples/kubernetes:v0.2.0 --namespace dev
  porter installation install --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter installation install MyAppInDev --file myapp/bundle.json
  porter installation install --parameter-set azure --param test-mode=true --param header-color=blue
  porter installation install --credential-set azure --credential-set kubernetes
  porter installation install --driver debug
  porter installation install --label env=dev --label owner=myuser

```

### Options

```
      --allow-docker-host-access     Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://getporter.org/configuration/#allow-docker-host-access for the full implications of this flag.
      --autobuild-disabled           Do not automatically build the bundle from source when the last build is out-of-date.
      --cnab-file string             Path to the CNAB bundle.json file.
  -c, --credential-set stringArray   Credential sets to use when running the bundle. It should be a named set of credentials and may be specified multiple times.
      --debug                        Run the bundle in debug mode.
  -d, --driver string                Specify a driver to use. Allowed values: docker, debug (default "docker")
  -f, --file porter.yaml             Path to the Porter manifest. Defaults to porter.yaml in the current directory.
      --force                        Force a fresh pull of the bundle
  -h, --help                         help for install
      --insecure-registry            Don't require TLS for the registry
  -l, --label strings                Associate the specified labels with the installation. May be specified multiple times.
  -n, --namespace string             Create the installation in the specified namespace. Defaults to the global namespace.
      --no-logs                      Do not persist the bundle execution logs
      --param stringArray            Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times.
  -p, --parameter-set stringArray    Parameter sets to use when running the bundle. It should be a named set of parameters and may be specified multiple times.
  -r, --reference string             Use a bundle in an OCI registry specified by the given reference.
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter installations](/cli/porter_installations/)	 - Installation commands

