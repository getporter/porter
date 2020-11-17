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

### Options

```
  -h, --help             help for build
      --name string      Override the bundle name
      --no-lint          Do not run the linter
  -v, --verbose          Enable verbose logging
      --version string   Override the bundle version
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

