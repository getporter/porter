---
title: "porter bundles lint"
slug: porter_bundles_lint
url: /cli/porter_bundles_lint/
---
## porter bundles lint

Lint a bundle

### Synopsis

Check the bundle for problems and adherence to best practices by running linters for porter and the mixins used in the bundle.

The lint command is run automatically when you build a bundle. The command is available separately so that you can just lint your bundle without also building it.

```
porter bundles lint [flags]
```

### Options

```
  -f, --file string     Path to the porter manifest file. Defaults to the bundle in the current directory.
  -h, --help            help for lint
  -o, --output string   Specify an output format.  Allowed values: plaintext, json (default "plaintext")
  -v, --verbose         Enable verbose logging
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter bundles](/cli/porter_bundles/)	 - Bundle commands

