---
title: "porter instances output show"
slug: porter_instances_output_show
url: /cli/porter_instances_output_show/
---
## porter instances output show

Show the output of a bundle instance

### Synopsis

Show the output of a bundle instance

```
porter instances output show NAME [--instance|-i INSTANCE] [flags]
```

### Examples

```
  porter instance output show kubeconfig
    porter instance output show subscription-id --instance azure-mysql
```

### Options

```
  -h, --help              help for show
  -i, --instance string   Specify the bundle instance to which the output belongs.
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter instances output](/cli/porter_instances_output/)	 - Output commands

