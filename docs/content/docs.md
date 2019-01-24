---
title: Porter Docs
description: All the magic of Porter explained
---

Porter takes the work out of creating CNAB bundles. It provides a declarative authoring 
experience that lets you to reuse existing bundles, and understands how to translate 
CNAB actions to Helm, Terraform, Azure, etc.

```console
$ porter create
created porter.yaml

$ porter build
created Dockerfile
created cnab/app/run

$ duffle build
```

Porter has both a buildtime cli, and a runtime cli. The buildtime handles (re)generating
a Dockerfile and copying files into the invocation image. The runtime handles
interpreting the porter.yaml file and executing the run actions (install, uninstall, etc).
