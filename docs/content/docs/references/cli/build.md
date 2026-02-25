---
title: "porter build"
slug: porter_build
url: /cli/porter_build/
---
## porter build

Build a bundle

### Synopsis

Builds the bundle in the current directory by generating a Dockerfile and a CNAB bundle.json, and then building the bundle image.

The docker driver builds the bundle image using the local Docker host. To use a remote Docker host, set the following environment variables:
  DOCKER_HOST (required)
  DOCKER_TLS_VERIFY (optional)
  DOCKER_CERT_PATH (optional)
'


```
porter build [flags]
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
      --build-arg stringArray       Set build arguments in the template Dockerfile (format: NAME=VALUE). May be specified multiple times. Max length is 5,000 characters.
      --build-context stringArray   Define additional build context with specified contents (format: NAME=PATH). May be specified multiple times. Porter automatically provides a 'porter-internal-userfiles' named context pointing to your bundle directory.
      --builder string              Set the name of the buildkit builder to use.
      --cache-from stringArray      Add cache source images to the build cache. May be specified multiple times.
      --cache-to stringArray        Add cache target images to the build cache.
      --custom stringArray          Define an individual key-value pair for the custom section in the form of NAME=VALUE. Use dot notation to specify a nested custom field. May be specified multiple times. Max length is 5,000 characters when used as a build argument.
  -d, --dir string                  Path to the build context directory where all bundle assets are located. Defaults to the current directory.
  -f, --file string                 Path to the Porter manifest. The path is relative to the build context directory. Defaults to porter.yaml in the current directory.
      --force                       Force a full rebuild from scratch, ignoring any cached data.
  -h, --help                        help for build
      --insecure-registry           Don't require TLS when pulling referenced images
      --name string                 Override the bundle name
      --no-cache                    Do not use the Docker cache when building the bundle image.
      --no-lint                     Do not run the linter
      --output string               Set docker output options (excluding type and name).
      --preserve-tags               Preserve the original tag name on referenced images
      --secret stringArray          Secret file to expose to the build (format: id=mysecret,src=/local/secret). Custom values are accessible as build arguments in the template Dockerfile and in the manifest using template variables. May be specified multiple times.
      --ssh stringArray             SSH agent socket or keys to expose to the build (format: default|<id>[=<socket>|<key>[,<key>]]). May be specified multiple times.
      --version string              Override the bundle version
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter](/cli/porter/)	 - With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://porter.sh/quickstart to learn how to use Porter.


