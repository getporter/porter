## Who is Porter For?
* [Personas](/personas)

## What is Porter?

Porter is a helper binary that [Duffle](https://github.com/deis/duffle) can build
into your CNAB invocation images. It provides a declarative authoring experience
for CNAB bundles that allows you to reuse existing bundles, and understands how to
translate CNAB actions to Helm, Terraform, Azure, etc.

```console
$ porter init
created porter.yaml

$ porter build
created Dockerfile
created cnab/app/run

$ duffle build
```

Porter has both a buildtime cli, and a runtime cli. The buildtime handles (re)generating
a Dockerfile and copying files into the invocation image. The runtime handles
interpreting the porter.yaml file and executing the run actions (install, uninstall, etc).
