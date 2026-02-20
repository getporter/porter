---
title: Parameters, Credentials, Outputs, and Images in Porter
description: How to wire parameters, credentials and outputs into steps
---

In the Porter manifest, you can declare both parameters and credentials. In addition to providing a mechanism for declaring parameters and credentials at the bundle level, Porter provides a way to declare how each of these are provided to mixins. This mechanism is also applicable to declaring how output from one mixin can be passed to another, as well as how to consume parameters, credentials and outputs from bundle dependencies. Finally, you can also use this technique to reference images defined in the `images` section of the manifest.

* [Wiring Installation Metadata](#wiring-installation-metadata)
* [Parameters](#parameters)
  * [File Parameters](#file-parameters)
  * [Wiring Parameters](#wiring-parameters)
* [Credentials](#credentials)
  * [Wiring Credentials](#wiring-credentials)
* [Outputs](#outputs)
  * [Wiring Outputs](#wiring-outputs)
* [Wiring Custom Metadata](#wiring-custom-metadata)
* [Wiring Images](#wiring-images)
* [Wiring Dependency Outputs](#wiring-dependency-outputs)
* [Combining References](#combining-references)

## Wiring Installation Metadata

Installation metadata is available at runtime via template variables and environment variables:

| Template variable      | Environment variable          | Description                              |
| ---------------------- | ----------------------------- | ---------------------------------------- |
| installation.name      | PORTER_INSTALLATION_NAME      | The name of the installation.            |
| installation.namespace | PORTER_INSTALLATION_NAMESPACE | The namespace of the installation.       |
| installation.id        | PORTER_INSTALLATION_ID        | A globally unique ID for the installation, stable across all runs of the same installation record. |

In the example below, we install a helm chart and set the release name to the installation name of the bundle:

```yaml
install:
  helm3:
    description: Install myapp
    name: ${ installation.name }
    chart: charts/myapp
```

Use `installation.id` when you need a value that is globally unique across Porter data stores and namespaces, for example to tag cloud resources so they can be traced back to a specific installation:

```yaml
install:
  exec:
    description: Create storage account
    command: az
    arguments:
      - storage
      - account
      - create
      - --name
      - myapp-storage
      - --tags
      - porter-installation-id=${ installation.id }
```

## Parameters

In order to declare a parameter in a Porter bundle, you first declare a parameters block with one or more parameters in the `porter.yaml`. For example, to declare a parameter named database_name, you might include the following YAML block:

```yaml
parameters:
- name: database_name
  type: string
```

This is the minimum required to create a parameter in Porter. Porter will specify an environment variable destination that defaults to the upper-cased name of the parameter.

You can also provide any other attributes, as specified by the CNAB [parameters](https://github.com/cnabio/cnab-spec/blob/master/101-bundle-json.md#parameters) specification. To specify a default value, for example, you could provide the following parameter definition:

```yaml
- name: database_name
  type: string
  default: "wordpress"
```

If you decide to use the default parameter field, you must set it as the empty value. 
However, it must be passed in as an empty type. For example, for an empty string, pass in `""` as the default, or for an object use `{}`:

```yaml
- name: command
  type: object
  default: {}
```

### File Parameters

Porter also enables the use of file parameters in a bundle.

For instance, a bundle might declare a parameter `mytar` of type `file`, to exist at `/cnab/app/mytar` in the execution environment:

```yaml
- name: mytar
  type: file
  path: /cnab/app/mytar
```

which can be used in a step like `install`:

```yaml
install:
  - exec:
      description: "Install"
      command: bash
      flags:
        c: tar zxvf /cnab/app/mytar
```

The syntax to pass a parameter to porter is the same for both regular and file parameters:

```console
$ porter install --param mytar=./my.tar.gz
```

See the [Parameters section of the Author Bundles doc](/docs/bundle/manifest/#parameters) for additional examples and configuration.

## Wiring Parameters

Once a parameter has been declared in the `porter.yaml`, Porter provides a simple mechanism for specifying how the parameter value should be wired into the mixin. To reference a parameter in the bundle, you use a notation of the form `${ bundle.parameters.PARAM_NAME }`. This can be used anywhere within step definition. For example, to use the `database_name` parameter defined above in the `set` block of the helm mixin:

```yaml
install:
- helm3:
    description: "Install MySQL"
    name: porter-ci-mysql
    chart: bitnami/mysql
    version: 6.14.2
    replace: true
    set:
      db.name: ${ bundle.parameters.database-name }
      db.user: "root"
```

Or to provide a parameter to the `command` attribute of the exec mixin:

```yaml
parameters:
  - name: command
    type: string
    default: "echo Hello World"

install:
  - description: "Install Hello World"
    exec:
      command: ${ bundle.parameters.command }
```

This syntax is used in dictionary, as above, or in a list:

```yaml
parameters:
  - name: command
    type: string
    default: "echo Hello World"

install:
  - description: "Install Hello World"
    exec:
      command: bash
      flags:
        c: ${ bundle.parameters.command }
```

NOTE: These references must be quoted, as in the examples above.

See [Parameters][parameters] to learn how parameters are passed in to Porter prior to bundle execution.


## Credentials

Credentials are defined in the `porter.yaml` with a YAML block of one more credential definitions. You can declare that a credential should be placed in a path within the bundle image or into an environment variable.

To declare a file injection:

```yaml
credentials:
- name: kubeconfig
  path: /home/nonroot/.kube/config
```

To declare an environment variable injection:

```yaml
credentials:
- name: SUBSCRIPTION_ID
  env: AZURE_SUBSCRIPTION_ID
```

## Wiring Credentials

The same mechanism for declaring how to use a parameter can be used for credentials. To declare a credential usage, references are defined with the following syntax: `${ bundle.credentials.CREDENTIAL_NAME}`.

When the bundle is executed, the Porter runtime will locate the parameter definition in the `porter.yaml` to determine where the parameter value has been stored. The Porter runtime will then rewrite the YAML block before it is passed to the mixin. See [Credentials][credentials] to learn how credentials work.

## Outputs

In addition to parameters and credentials, Porter introduces a type called an output. Outputs are values that are generated during the execution of a mixin. These could be things like a hostname for a newly provisioned Azure service or a generated password from something installed with Helm. These often need to be used in subsequent steps so Porter allows you to declare how to reference them after they have been created. The form used to declare these varies by mixin, but might look something like:

```yaml
install:
  - arm:
      description: "Create Azure MySQL"
      type: arm
      template: "arm/mysql.json"
      name: demo-mysql-azure-porter-demo-wordpress
      resourceGroup: "porter-test"
      parameters:
        administratorLogin: ${ bundle.parameters.mysql_user}
        administratorLoginPassword: ${ bundle.parameters.mysql_password }
        location: "eastus"
        serverName: "mysql-jeremy-porter-test-jan-2018"
        version: "5.7"
        sslEnforcement: "Disabled"
        databaseName: ${ bundle.parameters.database_name }
      outputs:
        - name: "MYSQL_URL"
          key: "MYSQL_HOST"
```

In this example, a new output will be created named `MYSQL_URL`. The Azure mixin allows you to specify the key to fetch the output from, in this case it is `MYSQL_HOST`. Each mixin can provide different ways of addressing outputs, so refer to the schema for each mixin. The Porter runtime will keep a map in memory with each of the outputs declared.

## Wiring Outputs

Once an output has been declared, it can be referenced in the same way as parameters and credentials. Outputs are referenced with the syntax `${ bundle.outputs.OUTPUT_NAME }`

For example, given the install step above, we can use the `MYSQL_URL` with the helm mixin in the following way:

```yaml
  - helm3:
      description: "Helm Install Wordpress"
      name: porter-ci-wordpress
      chart: bitnami/wordpress
      version: "9.9.3"
      set:
        mariadb.enabled: "false"
        externalDatabase.port: 3306
        externalDatabase.host: ${  bundle.outputs.MYSQL_URL }
        externalDatabase.user: ${ bundle.parameters.mysql_user }
        externalDatabase.password: ${ bundle.parameters.mysql_password }
        externalDatabase.database: ${ bundle.parameters.database_name }
```

Just like in the case of credentials and parameters, the value of the `bundle.outputs.MYSQL_URL` reference will be rewritten in the YAML before the helm mixin is invoked.

Parameters can also use the value from an output from the current bundle or one of its dependencies as its default value
using the `source` field when defining the parameter. This is useful for persisting data between bundle actions, such as
resource IDs created during install that are needed during upgrade or uninstall.

**Source an output from the current bundle**
```yaml
parameters:
- name: tfstate
  type: file
  path: /cnab/app/tfstate
  source:
    output: tfstate
```

**Source an output from a dependency**
```yaml
parameters:
- name: connection-string
  type: string
  source:
    dependency: mysql
    source: connstr
```

See [Persisting Data Between Bundle Actions](/docs/development/authoring-a-bundle/persisting-data/) for a complete guide
with working examples on when to use parameter sources versus state files.

## Wiring Custom Metadata

In the `porter.yaml`, you can define custom metadata and use it in your bundle.
These custom values are hard-coded into the bundle and cannot be modified at
runtime. If you need that, then you should use a parameter.

```yaml
custom:
  myApp:
    featureFlags:
      featureA: true
```

Now you can use the custom values in your actions like so:

```yaml
install:
  helm3:
    description: Install myapp
    chart: charts/myapp
    set:
      featureA: ${ bundle.custom.myApp.featureFlags.featureA }
```

## Wiring Images

In the `porter.yaml`, you can define what images will be used within the bundle with the `images` section:

```yaml
images:
  ALIAS:
    description: A very useful image
    imageType: docker
    repository: gcr.io/mcguffin-co/mcguffin
    digest: sha256:85b1a9
    tag: v1.1.0
```

These images will be used to build the `bundle.json` images section, but can also be referenced using the same syntax you would use for referencing `parameters`, `credentials`, and `outputs`.

```yaml
  - helm3:
      description: "Helm Install Wordpress"
      name: porter-ci-wordpress
      chart: bitnami/wordpress
      version: "9.9.3"
      set:
        image.repository: ${ bundle.images.ALIAS.repository }
        image.tag: ${ bundle.images.ALIAS.tag }
```

## Wiring Dependency Outputs

You can reference outputs from a dependency defined in your bundle using the syntax `${ bundle.dependencies.DEPENDENCY.outputs.OUTPUT }`.

For example, consider a bundle that creates a mysql defined with the following `porter.yaml`:

```yaml
name: mysql
version: 0.1.3
registry: getporter

mixins:
- helm3:
    repositories:
      bitnami:
        url: "https://charts.bitnami.com/bitnami"

credentials:
- name: kubeconfig
  path: /home/nonroot/.kube/config

parameters:
- name: database-name
  type: string
  default: mydb
  env: DATABASE_NAME
- name: mysql-user
  type: string
  env: MYSQL_USER

install:
- helm3:
    description: "Install MySQL"
    name: porter-ci-mysql
    chart: bitnami/mysql
    version: 6.14.2
    replace: true
    set:
      db.name: ${ bundle.parameters.database-name }
      db.user: ${ bundle.parameters.mysql-user }
    outputs:
    - name: mysql-root-password
      secret: porter-ci-mysql
      key: mysql-root-password
    - name: mysql-password
      secret: porter-ci-mysql
      key: mysql-password
```

With this bundle definition, we can build a second bundle to install wordpress and declare a dependency on this bundle. The `porter.yaml` for this might look something like:

```yaml
name: wordpress
version: 0.1.0
registry: getporter

mixins:
- helm3:
    repositories:
      bitnami:
        url: "https://charts.bitnami.com/bitnami"

dependencies:
  requires:
    - name: mysql
      bundle:
        reference: getporter/mysql:v0.1.3
      parameters:
        database_name: wordpress
        mysql_user: wordpress

credentials:
- name: kubeconfig
  path: /home/nonroot/.kube/config

parameters:
- name: wordpress-name
  type: string
  default: porter-ci-wordpress
  env: WORDPRESS_NAME
- name: wordpress-password
  type: string
  sensitive: true
  applyTo:
    - install
    - upgrade
- name: namespace
  type: string
  default: ''

install:
- helm3:
  description: "Install Wordpress"
  name: ${ bundle.parameters.wordpress-name }
  chart: bitnami/wordpress
  version: "9.9.3"  
  namespace: ${ bundle.parameters.namespace }
  replace: true
  set:
    wordpressPassword: ${ bundle.parameters.wordpress-password }
    externalDatabase.password: ${ bundle.dependencies.mysql.outputs.mysql-password }
    externalDatabase.port: 3306
    mariadb.enabled: false
  outputs:
    - name: wordpress-password
      secret: ${ bundle.parameters.wordpress-name }
      key: wordpress-password
```

The wordpress bundle declares a dependency on the `mysql` bundle, which we saw above. Now, we are able to refer to the parameters and the outputs from that bundle!

```yaml
install:
- helm3:
  description: "Install Wordpress"
  name: ${ bundle.parameters.wordpress-name }
  chart: bitnami/wordpress
  version: "9.9.3"      
  namespace: ${ bundle.parameters.namespace }
  replace: true
  set:
    wordpressPassword: ${ bundle.parameters.wordpress-password }
    externalDatabase.password: ${ bundle.dependencies.mysql.outputs.mysql-password }
    externalDatabase.port: 3306
    mariadb.enabled: false
```

For more information on how dependencies are handled, refer to the [dependencies](/docs/development/authoring-a-bundle/working-with-dependencies/) documentation.

## Combining References

It is possible to reference multiple parameters, credentials and/or outputs in a single place. You can combine the expressions as follows:

```yaml
install:
- helm3:
    description: "Install Java App"
    name: ${ bundle.parameters.cool-app}
    chart: bitnami/wordpress
    version: "9.9.3"
    replace: true
    set:
      jdbc_url: "jdbc:mysql://${ bundle.outputs.mysql_host }:${ bundle.outputs.mysql_port }/${ bundle.parameters.database_name }
```

[credentials]: /docs/introduction/concepts-and-components/intro-credentials/
[parameters]: /docs/introduction/concepts-and-components/intro-parameters/
