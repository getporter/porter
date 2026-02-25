---
title: "porter lint"
slug: porter_lint
url: /cli/porter_lint/
---
## porter lint

Lint a bundle

### Synopsis

Check the bundle for problems and adherence to best practices by running linters for porter and the mixins used in the bundle.

The lint command is run automatically when you build a bundle. The command is available separately so that you can just lint your bundle without also building it.

```
porter lint [flags]
```

### Examples

```
  porter lint
  porter lint --file path/to/porter.yaml
  porter lint --output plaintext

```

### Options

```
  -f, --file string     Path to the porter manifest file. Defaults to the bundle in the current directory.
  -h, --help            help for lint
  -o, --output string   Specify an output format.  Allowed values: plaintext, json (default "plaintext")
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


