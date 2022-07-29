---
title: "porter version"
slug: porter_version
url: /cli/porter_version/
---
## porter version

Print the application version

```
porter version [flags]
```

### Options

```
  -h, --help            help for version
  -o, --output string   Specify an output format.  Allowed values: json, plaintext (default "plaintext")
  -s, --system          Print system debug information
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


