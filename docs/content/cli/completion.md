---
title: "porter completion"
slug: porter_completion
url: /cli/porter_completion/
---
## porter completion

Generate completion script

### Synopsis

Capture the output of the completion command to a file for your shell environment.

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
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

