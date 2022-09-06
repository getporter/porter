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
  porter parameters generate
  porter parameters generate myparamset --reference getporter/hello-llama:v0.1.1 --namespace dev
  porter parameters generate myparamset --label owner=myname --reference getporter/hello-llama:v0.1.1
  porter parameters generate myparamset --reference localhost:5000/getporter/hello-llama:v0.1.1 --insecure-registry --force
  porter parameters generate myparamset --file myapp/porter.yaml
  porter parameters generate myparamset --cnab-file myapp/bundle.json

```

### Options

```
      --cnab-file string    Path to the CNAB bundle.json file.
  -f, --file string         Path to the porter manifest file. Defaults to the bundle in the current directory.
      --force               Force a fresh pull of the bundle
  -h, --help                help for generate
      --insecure-registry   Don't require TLS for the registry
  -l, --label strings       Associate the specified labels with the parameter set. May be specified multiple times.
  -n, --namespace string    Namespace in which the parameter set is defined. Defaults to the global namespace.
  -r, --reference string    Use a bundle in an OCI registry specified by the given reference.
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter parameters](/cli/porter_parameters/)	 - Parameter set commands

