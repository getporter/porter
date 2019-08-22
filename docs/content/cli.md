---
title: Porter Commands
description: Porter CLI Commands Explained
---

* [Bundle Commands](#bundle-commands)
* [Instance Commands](#instance-commands)
* [Instance Outputs Commands](#outputs-commands)
* [Mixin Commands](#mixin-commands)
* [Meta Commands](#meta-commands)

## Bundle Commands

### Create

This command is aliased and is available both as `porter create`
and `porter bundle create`.

```console
 $ porter create --help
Create a bundle. This generates a porter manifest, porter.yaml, and the CNAB run script in the current directory.

Usage:
  porter create [flags]

Flags:
  -h, --help   help for create

Global Flags:
      --debug   Enable debug logging
```

### Build

This command is aliased and is available both as `porter build`
and `porter bundle build`.

```console
 $ porter build --help
Builds the bundle in the current directory by generating a Dockerfile and a CNAB bundle.json, and then building the invocation image.

Usage:
  porter build [flags]

Flags:
  -h, --help   help for build

Global Flags:
      --debug   Enable debug logging
```

### Install

This command is aliased and is available both as `porter install`
and `porter bundle install`.

```console
$ porter install --help
Install a bundle.

The first argument is the bundle instance name to create for the installation.
This defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's
invocation image, but an alternate driver may be supplied via '--driver/-d'. For
instance, the 'debug' driver may be specified, which simply logs the info given
to it and then exits.

Usage:
  porter install [INSTANCE] [flags]

Examples:
  porter install
  porter install --insecure
  porter install MyAppInDev --file myapp/bundle.json
  porter install --param-file base-values.txt --param-file dev-values.txt --param test-mode=true --param header-color=blue
  porter install --cred azure --cred kubernetes
  porter install --driver debug


Flags:
  -c, --cred strings         Credential to use when installing the bundle. May be 
                             either a named set of credentials or a filepath, and 
                             specified multiple times.
  -d, --driver string        Specify a driver to use. Allowed values: docker, debug 
                             (default "docker")
  -f, --file string          Path to the CNAB definition to install. Defaults to the 
                             bundle in the current directory.
  -h, --help                 help for install
      --insecure             Allow working with untrusted bundles
      --param strings        Define an individual parameter in the form NAME=VALUE. 
                             Overrides parameters set with the same name using --param-file. 
                             May be specified multiple times.
      --param-file strings   Path to a parameters definition file for the bundle, 
                             each line in the form of NAME=VALUE. May be specified 
                             multiple times.

Global Flags:
      --debug   Enable debug logging
```

### Upgrade 

This command is aliased and is available both as `porter upgrade` and `porter
bundle upgrade`.

```console
$ porter upgrade --help
Upgrade a bundle.

The first argument is the bundle instance name to upgrade. This defaults to the
name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's
invocation image, but an alternate driver may be supplied via '--driver/-d'. For
instance, the 'debug' driver may be specified, which simply logs the info given
to it and then exits.

Usage:
  porter upgrade [INSTANCE] [flags]

Examples:
  porter upgrade
  porter upgrade --insecure
  porter upgrade MyAppInDev --file myapp/bundle.json
  porter upgrade --param-file base-values.txt --param-file dev-values.txt --param test-mode=true --param header-color=blue
  porter upgrade --cred azure --cred kubernetes
  porter upgrade --driver debug


Flags:
  -c, --cred strings         Credential to use when installing the bundle. May be 
                             either a named set of credentials or a filepath, and
                             specified multiple times.
  -d, --driver string        Specify a driver to use. Allowed values: docker,
                             debug (default "docker")
  -f, --file string          Path to the CNAB definition to upgrade. Defaults to
                             the bundle in the current directory.
  -h, --help                 help for upgrade
      --insecure             Allow working with untrusted bundles
      --param strings        Define an individual parameter in the form NAME=VALUE.
                             Overrides parameters set with the same name using 
                             --param-file. May be specified multiple times.
      --param-file strings   Path to a parameters definition file for the bundle,
                             each line in the form of NAME=VALUE. May be specified
                             multiple times.

Global Flags:
      --debug   Enable debug logging
```

### Invoke

This command is aliased and is available both as `porter invoke`
and `porter bundle invoke`.

```console
Invoke a custom action on a bundle.

The first argument is the bundle instance name upon which to invoke the action.
This defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's
invocation image, but an alternate driver may be supplied via '--driver/-d'. For
instance, the 'debug' driver may be specified, which simply logs the info given
to it and then exits.

Usage:
  porter invoke [INSTANCE] --action ACTION [flags]

Examples:
  porter invoke --action ACTION
  porter invoke --action ACTION MyAppInDev --file myapp/bundle.json
  porter invoke --action ACTION --param-file base-values.txt --param-file dev-values.txt --param test-mode=true --param header-color=blue
  porter invoke --action ACTION --cred azure --cred kubernetes
  porter invoke --action ACTION --driver debug
  porter invoke --action ACTION MyAppFromTag --tag deislabs/porter-kube-bundle:v1.0


Flags:
      --action string        Custom action name to invoke.
      --cnab-file string     Path to the CNAB bundle.json file.
  -c, --cred strings         Credential to use when installing the bundle. May be 
                             either a named set of credentials or a filepath, and 
                             specified multiple times.
  -d, --driver string        Specify a driver to use. Allowed values: docker, debug 
                             (default "docker")
  -f, --file string          Path to the porter manifest file. Defaults to the bundle 
                             in the current directory.
  -h, --help                 help for invoke
      --insecure-registry    Don't require TLS for the registry
      --param strings        Define an individual parameter in the form NAME=VALUE.
                             Overrides parameters set with the same name using --param-file.
                             May be specified multiple times.
      --param-file strings   Path to a parameters definition file for the bundle, each line
                             in the form of NAME=VALUE. May be specified multiple times.
  -t, --tag string           Use a bundle in an OCI registry specified by the given tag

Global Flags:
      --debug   Enable debug logging
```

### Uninstall

This command is aliased and is available both as `porter uninstall`
and `porter bundle uninstall`.

```console
$ porter uninstall --help
Uninstall a bundle

The first argument is the bundle instance name to uninstall. This defaults to
the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's
invocation image, but an alternate driver may be supplied via '--driver/-d'. For
instance, the 'debug' driver may be specified, which simply logs the info given
to it and then exits.

Usage:
  porter uninstall [INSTANCE] [flags]

Examples:
  porter uninstall
  porter uninstall --insecure
  porter uninstall MyAppInDev --file myapp/bundle.json
  porter uninstall --param-file base-values.txt --param-file dev-values.txt --param test-mode=true --param header-color=blue
  porter uninstall --cred azure --cred kubernetes
  porter uninstall --driver debug


Flags:
  -c, --cred strings         Credential to use when uninstalling the bundle. May 
                             be either a named set of credentials or a filepath, 
                             and specified multiple times.
  -d, --driver string        Specify a driver to use. Allowed values: docker, debug (default "docker")
  -f, --file string          Path to the CNAB definition to uninstall. Defaults to 
                             the bundle in the current directory. Optional unless a 
                             newer version of the bundle should be used to uninstall 
                             the bundle.
  -h, --help                 help for uninstall
      --insecure             Allow working with untrusted bundles
      --param strings        Define an individual parameter in the form NAME=VALUE. 
                             Overrides parameters set with the same name using --param-file. 
                             May be specified multiple times.
      --param-file strings   Path to a parameters definition file for the bundle, each line 
                             in the form of NAME=VALUE. May be specified multiple times.

Global Flags:
      --debug   Enable debug logging
```

## Instance Commands

### Instances List

This command is available both as `porter instance list` and `porter list`.

```console
$ porter instance list --help
List all bundles installed by Porter.

A listing of bundles currently installed by Porter will be provided, along with
metadata such as creation time, last action, last status, etc.

Optional output formats include json and yaml.

Usage:
  porter bundles list [flags]

Examples:
  porter instance list
  porter instance list -o json

Flags:
  -h, --help            help for list
  -o, --output string   Specify an output format.
                        Allowed values: table, json, yaml (default "table")

Global Flags:
      --debug   Enable debug logging
```

### Instance Show

This command is available both as `porter instance show` and `porter show`.

```console
$ porter instance show --help
Displays info relating to a bundle instance, including status and a listing of outputs.

Usage:
  porter bundles show [INSTANCE] [flags]

Examples:
  porter bundle show [INSTANCE]

Optional output formats include json and yaml.


Flags:
  -h, --help            help for show
  -o, --output string   Specify an output format.  Allowed values: table, json, yaml (default "table")

Global Flags:
      --debug   Enable debug logging
```

## Instance Outputs Commands

### Instance Outputs List

Outputs from a bundle instance can be listed via `porter instance outputs list`.

```console
 $ porter instance outputs list --help
Displays a listing of bundle outputs.

Usage:
  porter bundles output list [INSTANCE] [flags]

Examples:
  porter instance outputs list
  porter instance outputs list --instance another-bundle

Optional output formats include json and yaml.


Flags:
  -i, --instance string   Specify the instance to which the output belongs.
  -h, --help            help for list
  -o, --output string   Specify an output format.  Allowed values: table, json, yaml (default "table")

Global Flags:
      --debug   Enable debug logging
```

### Instance Outputs Show

An individual output from a bundle instance can be retrieved via `porter instance output show`.

```console
 $ porter instance outputs show --help
Show a bundle output.

Usage:
  porter bundles output show NAME [--instance|-i INSTANCE] [flags]

Examples:
  porter instance output show kubeconfig
  porter instance output show subscription-id --instance azure-mysql

Flags:
  -i, --instance string   Specify the instance to which the output belongs.
  -h, --help              help for show

Global Flags:
      --debug   Enable debug logging
```

## Mixin Commands

### Mixins List

```console
$ porter mixins list --help
List installed mixins

Usage:
  porter mixins list [flags]

Flags:
  -h, --help            help for list
  -o, --output string   Output format, allowed values are: table, json (default "table")

Global Flags:
      --debug   Enable debug logging
```

### Mixins Feed Template

```console
$ porter mixins feed template --help
Create an atom feed template in the current directory

Usage:
  porter mixins feed template [flags]

Flags:
  -h, --help   help for template

Global Flags:
      --debug   Enable debug logging
```

### Mixins Feed Generate 

```console
$ porter mixins feed generate --help
Generate an atom feed from the mixins in a directory.

A template is required, providing values for text properties such as the author
name, base URLs and other values that cannot be inferred from the mixin file
names. You can make a default template by running 'porter mixins feed template'.

The file names of the mixins must follow the naming conventions required of
published mixins:

VERSION/MIXIN-GOOS-GOARCH[FILE_EXT]

More than one mixin may be present in the directory, and the directories may be
nested a few levels deep, as long as the file path ends with the above naming
convention, porter will find and match it. Below is an example directory
structure that porter can list to generate a feed:

bin/
└── v1.2.3/
    ├── mymixin-darwin-amd64
    ├── mymixin-linux-amd64
    └── mymixin-windows-amd64.exe

See https://porter.sh/mixin-distribution more details.

Usage:
  porter mixins feed generate [flags]

Examples:
  porter mixin feed generate
  porter mixin feed generate --dir bin --file bin/atom.xml --template porter-atom-template.xml

Flags:
  -d, --dir string        The directory to search for mixin versions to publish in
                          the feed. Defaults to the current directory.
  -f, --file string       The path of the atom feed output by this command. (default "atom.xml")
  -h, --help              help for generate
  -t, --template string   The template atom file used to populate the text fields 
                          in the generated feed. (default "atom-template.xml")

Global Flags:
      --debug   Enable debug logging
```

## Meta Commands

### Schema

```schema 
$ porter schema --help
Print the JSON schema for the Porter manifest

Usage:
  porter schema [flags]

Flags:
  -h, --help   help for schema

Global Flags:
      --debug   Enable debug logging
```

### Version

```console
$ porter version --help
Print the application version

Usage:
  porter version [flags]

Flags:
  -h, --help   help for version

Global Flags:
      --debug   Enable debug logging
```