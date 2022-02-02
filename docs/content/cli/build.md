---
title: "porter build"
slug: porter_build
url: /cli/porter_build/
---
## porter build

Build a bundle

### Synopsis

Builds the bundle in the current directory by generating a Dockerfile and a CNAB bundle.json, and then building the invocation image.

Porter uses the docker driver as the default build driver, an alternate driver may be supplied via --driver or the PORTER_BUILD_DRIVER environment variable.


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

```

### Options

```
  -d, --dir string         Path to the build context directory where all bundle assets are located.
      --driver string      Experimental. Driver for building the invocation image. Allowed values are: docker, buildkit (default "docker")
  -f, --file porter.yaml   Path to the Porter manifest. Defaults to porter.yaml in the current directory.
  -h, --help               help for build
      --name string        Override the bundle name
      --no-lint            Do not run the linter
  -v, --verbose            Enable verbose logging
      --version string     Override the bundle version
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter](/cli/porter/)	 - With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://porter.sh/quickstart to learn how to use Porter.


