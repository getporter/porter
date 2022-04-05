---
title: "porter build"
slug: porter_build
url: /cli/porter_build/
---
## porter build

Build a bundle

### Synopsis

Builds the bundle in the current directory by generating a Dockerfile and a CNAB bundle.json, and then building the invocation image.

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
      --build-arg stringArray   Set build arguments in the template Dockerfile (format: NAME=VALUE). May be specified multiple times.
  -d, --dir string              Path to the build context directory where all bundle assets are located.
  -f, --file porter.yaml        Path to the Porter manifest. Defaults to porter.yaml in the current directory.
  -h, --help                    help for build
      --name string             Override the bundle name
      --no-cache                Do not use cache when building the image.
      --no-lint                 Do not run the linter
      --secret stringArray      Secret file to expose to the build (format: id=mysecret,src=/local/secret). May be specified multiple times.
      --ssh stringArray         SSH agent socket or keys to expose to the build (format: default|<id>[=<socket>|<key>[,<key>]]). May be specified multiple times.
  -v, --verbose                 Enable verbose logging
      --version string          Override the bundle version
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


