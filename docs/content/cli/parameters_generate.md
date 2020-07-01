---
title: "porter parameters generate"
slug: porter_parameters_generate
url: /cli/porter_parameters_generate/
---
## porter parameters generate

Generate Parameter Set

### Synopsis

Generate a named set of parameters.

The first argument is the name of parameter set you wish to generate. If not
provided, this will default to the bundle name. By default, Porter will
generate a parameter set for the bundle in the current directory. You may also
specify a bundle with --file.

Bundles define 1 or more parameter(s) that are required to interact with a
bundle. The bundle definition defines where the parameter should be delivered
to the bundle, i.e. via DB_USERNAME. A parameter set, on the other hand,
represents the source data that you wish to use when interacting with the
bundle. These will typically be environment variables or files on your local
file system.

When you wish to install, upgrade or delete a bundle, Porter will use the
parameter set to determine where to read the necessary information from and
will then provide it to the bundle in the correct location. 

```
porter parameters generate [NAME] [flags]
```

### Examples

```
  porter parameter generate
  porter parameter generate myparamset --file myapp/porter.yaml
  porter parameter generate myparamset --tag getporter/porter-hello:v0.1.0
  porter parameter generate myparamset --cnab-file myapp/bundle.json --dry-run

```

### Options

```
      --cnab-file string   Path to the CNAB bundle.json file.
  -f, --file string        Path to the porter manifest file. Defaults to the bundle in the current directory.
      --force              Force a fresh pull of the bundle
  -h, --help               help for generate
      --tag string         Use a bundle in an OCI registry specified by the given tag.
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter parameters](/cli/porter_parameters/)	 - Parameter set commands

