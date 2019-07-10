---
title: Authoring Bundles
description: Authoring Bundles with Porter
---

Porter generates a bundle from its manifest, porter.yaml. The manifest is made up of a few components:

* [Bundle Metadata](#bundle-metadata)
* [Mixins](#mixins)
* [Parameters](#parameters)
* [Outputs](#outputs)
* [Credentials](#credentials)
* [Bundle Actions](#bundle-actions)
* [Dependencies](#dependencies)
* [Image Map](#image-map)
* [Generated Files](#generated-files)

We have full [examples](https://github.com/deislabs/porter/tree/master/examples) of Porter manifests in the Porter repository.

## Bundle Metadata

A lot of the metadata is defined by the [CNAB Spec](https://github.com/deislabs/cnab-spec/blob/master/101-bundle-json.md)
although Porter does have extra fields that are specific to making Porter bundles.

```yaml
name: porter-azure-wordpress
description: Install Wordpress on Azure
version: 0.1.0
invocationImage: deislabs/porter-azure-wordpress:latest
dockerfile: dockerfile.tmpl
```

* `name`: The name of the bundle
* `description`: A description of the bundle
* `version`: The version of the bundle, uses [semver](https://semver.org)
* `invocationImage`: The name of the container image to tag the bundle with when it is built. The format is
`REGISTRY/IMAGE:TAG`. Porter will push to this location during `porter publish` so select a location that you have access to.
* `dockerfile`: OPTIONAL. The relative path to a Dockerfile to use as a template during `porter build`. It is your responsibility
    to provide a suitable base image, for example one that has root ssl certificates installed. When a Dockerfile template is
    not specified, Porter automatically copies the contents of the current directory into `$BUNDLE_DIR` of the invocation image. 
    When using a Dockerfile template, you must manually copy any files you need in your bundle using COPY statements.

## Mixins

Mixins are adapters between the Porter and an existing tool or system. They know how to talk to Porter to include everything
they need to run, such as a CLI or config files, and how to execute their steps in the Porter manifest.

Anyone can [create a mixin](/mixin-dev-guide/), here's a list of the mixins that are installed with Porter by default:

* exec - run shell scripts and commands
* helm - use the helm cli
* azure - provision services on the Azure cloud

Declare the mixins that your bundle uses with the `mixins` section of the manifest:

```yaml
mixins:
- exec
- helm
```

See [Using Mixins](/using-mixins) to learn more about how mixins work.

## Parameters

Parameters are part of the [CNAB Spec](https://github.com/deislabs/cnab-spec/blob/master/101-bundle-json.md#parameters) and
allow you to pass in configuration values when you execute the bundle.

```yaml
parameters:
- name: mysql_user
  type: string
  default: azureuser
- name: mysql_password
  type: string
  sensitive: true
- name: database_name
  type: string
  default: "wordpress"
  destination:
    env: MYSQL_DATABASE
```

* `name`: The name of the parameter.
* `type`: The data type of the parameter: string, integer, number, boolean.
* `destination`: The destination in the bundle to define the parameter.
  * `env`: The name for the environment variable. Defaults to the name of the parameter in upper case.
  * `path`: The path for the file. Required for file paths, there is no default.
* `sensitive`: Optional. Designate this parameter's value as sensitive, for masking in console output.
 
## Outputs

Outputs are part of the [CNAB Spec](https://github.com/deislabs/cnab-spec/blob/master/101-bundle-json.md#outputs) and
allow outputs generated during the course of executing a bundle to be accessed.  These are global/bundle-wide outputs,
as opposed to step outputs described in [Parameters, Credentials and Outputs](/wiring/).  However, as of writing, each
bundle output is only valid if it references a step output declared under one or more [bundle actions](#bundle-actions).

```yaml
outputs:
- name: mysql_user
  type: string
  description: "MySQL user name"
- name: mysql_password
  type: string
  applyTo:
    - "install"
    - "upgrade"
```

* `name`: The name of the output.
* `type`: The data type of the output: string, integer, number, boolean.
* `applyTo`: (Optional) Restrict this output to a given list of actions. If empty or missing, applies to all actions.
* `description`: (Optional) A brief description of the given output.
* `sensitive`: (Optional) Designate an output as sensitive. Defaults to false.

### Definition Schema for Parameters and Outputs

The [CNAB Spec for definitions](https://github.com/deislabs/cnab-spec/blob/master/101-bundle-json.md#definitions)
applies to both parameters and outputs.  It has full support for [json schema](https://json-schema.org). We aren't
quite there yet but are working towards it. Here is what Porter supports for defining schema for parameters and outputs
currently:

* `default`: The default value for the parameter. When a default is not provided, the `required` is defaulted to true.
* `required`: Specify if the parameter must be specified when the bundle is executed.
* `enum`: A list of allowed values.
* [Numeric Range](https://json-schema.org/understanding-json-schema/reference/numeric.html?highlight=minimum#range) 
    using `minimum`, `maximum`, `exclusiveMinimum` and `exclusiveMaximum`.
* String length using `minLength` and `maxLength`.

## Credentials

Credentials are part of the [CNAB Spec](https://github.com/deislabs/cnab-spec/blob/master/802-credential-sets.md) and allow
you to pass in sensitive data when you execute the bundle, such as passwords or configuration files.

When the bundle is executed, for example when you run `porter install`, the installer will look on your local system
for the named credential and then place the value or file found in the bundle as either an environment variable or file.

By default, all credential values are considered sensitive and will be masked in console output.

```yaml
credentials:
- name: SUBSCRIPTION_ID
  env: AZURE_SUBSCRIPTION_ID
- name: CLIENT_ID
  env: AZURE_CLIENT_ID
- name: TENANT_ID
  env: AZURE_TENANT_ID
- name: CLIENT_SECRET
  env: AZURE_CLIENT_SECRET
- name: kubeconfig
  path: /root/.kube/config
```

* `name`: The name of the credential on the local system.
* `env`: The name of the environment variable to create with the value from the credential.
* `path`: The file path to create with the file from the credential.

## Bundle Actions

Porter supports the three CNAB actions: install, upgrade, and uninstall. Within each action, you define an ordered list
of steps to execute. Each step has a mixin, a `description`, and optionally `outputs`.

```yaml
install:
- description: "Install MySQL"
  helm:
    name: mydb
    chart: stable/mysql
    version: 0.10.2
    set:
      mysqlDatabase: "{{ bundle.parameters.database-name }}"
      mysqlUser: "{{ bundle.parameters.mysql-user }}"
  outputs:
  - name: mysql-root-password
    secret: mydb-creds
    key: mysql-root-password
  - name: mysql-password
    secret: mydb-creds
    key: mysql-password
```

* `description`: A description of the step, used for logging.
* `MIXIN`: The name of the mixin that will handle this step. In the example above, `helm` is the mixin.
* `outputs`: Any outputs provided by the steps. The `name` is required but the rest of the the schema for the 
output is specific to the mixin. In the example above, the mixin will make the Kubernetes secret data available as outputs.
By default, all output values are considered sensitive and will be masked in console output.

## Dependencies

See [dependencies](/dependencies/) for more details on how Porter handles dependencies.

```yaml
dependencies:
- name: mysql
  parameters:
    database_name: wordpress
    mysql_user: wordpress
```

* `name`: The name of the bundle.
* `parameters`: Optionally set default values for parameters in the bundle.

## Image Map

The Image Map is part of the [CNAB Spec](https://github.com/deislabs/cnab-spec/blob/master/103-bundle-runtime.md#image-maps).
The `imageMap` data from the manifest is made available at runtime `/cnab/app/image-map.json` where you may access it
from a script. 

Note: Neither porter nor the DeisLabs mixins use this information.

## Generated Files

In addition to the porter manifest, Porter generates a few files for you to create a compliant CNAB Spec bundle.

* `cnab/`: This directory structure is created during `porter init`. You can add files to this directory and they will
be copied into the final bundle so that you can access them at runtime. The path to this directory at runtime is `/cnab`.
* `cnab/app/run`: This file is created during `porter init` and should not be modified.
* `Dockerfile`: This file is generated during `porter build` and cannot be modified.

## See Also

* [Using Mixins](/using-mixins/)
* [Bundle Dependencies](/dependencies/)
* [Parameters, Credentials and Outputs](/wiring/)
