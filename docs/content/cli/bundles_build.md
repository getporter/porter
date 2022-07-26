---
title: "porter bundles build"
slug: porter_bundles_build
url: /cli/porter_bundles_build/
---
## porter bundles build

Build a bundle

### Synopsis

Builds the bundle in the current directory by generating a Dockerfile and a CNAB bundle.json, and then building the invocation image.

The docker driver builds the bundle image using the local Docker host. To use a remote Docker host, set the following environment variables:
  DOCKER_HOST (required)
  DOCKER_TLS_VERIFY (optional)
  DOCKER_CERT_PATH (optional)
'


```
porter bundles build [flags]
```

### Examples

```
  porter build
  porter build --name newbuns
  porter build --version 0.1.0
  porter build --file path/to/porter.yaml
  porter build --dir path/to/build/context
  porter build --custom version=0.2.0 --custom myapp.version=0.1.2

```

### Options

```
      --build-arg stringArray   Set build arguments in the template Dockerfile (format: NAME=VALUE). May be specified multiple times.
      --custom stringArray      Define an individual key-value pair for the custom section in the form of NAME=VALUE. Use dot notation to specify a nested custom field. May be specified multiple times.
  -d, --dir string              Path to the build context directory where all bundle assets are located. Defaults to the current directory.
  -f, --file string             Path to the Porter manifest. The path is relative to the build context directory. Defaults to porter.yaml in the current directory.
  -h, --help                    help for build
      --name string             Override the bundle name
      --no-cache                Do not use the Docker cache when building the bundle's invocation image.
      --no-lint                 Do not run the linter
      --secret stringArray      Secret file to expose to the build (format: id=mysecret,src=/local/secret). Custom values are assessible as build arguments in the template Dockerfile and in the manifest using template variables. May be specified multiple times.
      --ssh stringArray         SSH agent socket or keys to expose to the build (format: default|<id>[=<socket>|<key>[,<key>]]). May be specified multiple times.
      --version string          Override the bundle version
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter bundles](/cli/porter_bundles/)	 - Bundle commands

