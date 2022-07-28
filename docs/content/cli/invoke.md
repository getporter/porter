---
title: "porter invoke"
slug: porter_invoke
url: /cli/porter_invoke/
---
## porter invoke

Invoke a custom action on an installation

### Synopsis

Invoke a custom action on an installation.

The first argument is the installation name upon which to invoke the action. This defaults to the name of the bundle.

Porter uses the docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d' or the PORTER_RUNTIME_DRIVER environment variable.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.

The docker driver runs the bundle container using the local Docker host. To use a remote Docker host, set the following environment variables:
  DOCKER_HOST (required)
  DOCKER_TLS_VERIFY (optional)
  DOCKER_CERT_PATH (optional)


```
porter invoke [INSTALLATION] --action ACTION [flags]
```

### Examples

```
  porter invoke --action ACTION
  porter invoke --reference ghcr.io/getporter/examples/kubernetes:v0.2.0
  porter invoke --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter invoke --action ACTION MyAppInDev --file myapp/bundle.json
  porter invoke --action ACTION  --parameter-set azure --param test-mode=true --param header-color=blue
  porter invoke --action ACTION --credential-set azure --credential-set kubernetes
  porter invoke --action ACTION --driver debug

```

### Options

```
      --action string                Custom action name to invoke.
      --allow-docker-host-access     Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://getporter.org/configuration/#allow-docker-host-access for the full implications of this flag.
      --cnab-file string             Path to the CNAB bundle.json file.
  -c, --credential-set stringArray   Credential sets to use when running the bundle. It should be a named set of credentials and may be specified multiple times.
      --debug                        Run the bundle in debug mode.
  -d, --driver string                Specify a driver to use. Allowed values: docker, debug (default "docker")
  -f, --file string                  Path to the porter manifest file. Defaults to the bundle in the current directory.
      --force                        Force a fresh pull of the bundle
  -h, --help                         help for invoke
      --insecure-registry            Don't require TLS for the registry
  -n, --namespace string             Namespace of the specified installation. Defaults to the global namespace.
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

* [porter](/cli/porter/)	 - With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://getporter.org/quickstart to learn how to use Porter.


