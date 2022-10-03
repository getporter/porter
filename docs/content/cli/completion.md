---
title: "porter completion"
slug: porter_completion
url: /cli/porter_completion/
---
## porter completion

Generate completion script

### Synopsis

Save the output of this command to a file and load the file into your shell.
For additional details see: https://getporter.org/install#command-completion

```
porter completion [bash|zsh|fish|powershell]
```

### Examples

```
porter completion bash > /usr/local/etc/bash_completions.d/porter
```

### Options

```
  -h, --help   help for completion
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


