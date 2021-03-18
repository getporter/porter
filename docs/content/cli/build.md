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
  porter build --dir path/to/bundle

```

### Options

```
  -d, --dir string         Path to the build context directory where all bundle assets are located.
  -f, --file porter.yaml   Path to the Porter manifest. Defaults to porter.yaml in the current directory.
  -h, --help               help for build
      --name string        Override the bundle name
      --no-lint            Do not run the linter
  -v, --verbose            Enable verbose logging
      --version string     Override the bundle version
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

