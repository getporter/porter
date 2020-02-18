---
title: Parameters, Credentials, Outputs, and Images in Porter
description: How to wire parameters, credentials and outputs into steps
---

In the Porter manifest, you can declare both parameters and credentials. In addition to providing a mechanism for declaring parameters and credentials at the bundle level, Porter provides a way to declare how each of these are provided to [mixins][mixin-architecture]. This mechanism is also applicable to declaring how output from one mixin can be passed to another, as well as how to consume parameters, credentials and outputs from bundle dependencies. Finally, you can also use this technique to reference images defined in the `images` section of the manifest.

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

### File Parameters

Porter also enables the use of file parameters in a bundle.

For instance, a bundle might declare a parameter `mytar` of type `file`, to exist at `/root/mytar` in the runtime container:

```yaml
- name: mytar
  type: file
  path: /root/mytar
```

which can be used in a step like `install`:

```yaml
install:
  - exec:
      description: "Install"
      command: bash
      flags:
        c: tar zxvf /root/mytar
```

Passing in the local file representing the `mytar` parameter is done similarly to any other parameter:

```console
$ porter install --param mytar=./my.tar.gz
```

## Wiring Parameters

Once a parameter has been declared in the `porter.yaml`, Porter provides a simple mechanism for specifying how the parameter value should be wired into the mixin. To reference a parameter in the bundle, you use a notation of the form `"{{ bundle.parameters.PARAM_NAME }}"`. This can be used anywhere within step definition. For example, to use the `database_name` parameter defined above in the `set` block of the helm mixin:

```yaml
install:
- helm:
    description: "Install MySQL"
    name: porter-ci-mysql
    chart: stable/mysql
    version: 0.10.2
    replace: true
    set:
      mysqlDatabase: "{{ bundle.parameters.database-name }}"
      mysqlUser: "root"
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
      command: "{{ bundle.parameters.command }}"
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
        c: "{{ bundle.parameters.command }}"
```

NOTE: These references must be quoted, as in the examples above.


## Credentials

Credentials are defined in the `porter.yaml` with a YAML block of one more credential definitions. You can declare that a credential should be placed in a path within the invocation image or into an environment variable.

To declare a file injection:

```yaml
credentials:
- name: kubeconfig
  path: /root/.kube/config
```

To declare an environment variable injection:

```yaml
credentials:
- name: SUBSCRIPTION_ID
  env: AZURE_SUBSCRIPTION_ID
```

## Wiring Credentials

The same mechanism for declaring how to use a parameter can be used for credentials. To declare a credential usage, references are defined with the following syntax: `"{{ bundle.credentials.CREDENTIAL_NAME}}"`.

When the bundle is executed, the Porter runtime will locate the parameter definition in the `porter.yaml` to determine where the parameter value has been stored. The Porter runtime will then rewrite the YAML block before it is passed to the mixin. To understand how credentials work, see [how credentials work][how-credentials-work] page.

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
        administratorLogin: "{{ bundle.parameters.mysql_user}}"
        administratorLoginPassword: "{{ bundle.parameters.mysql_password }}"
        location: "eastus"
        serverName: "mysql-jeremy-porter-test-jan-2018"
        version: "5.7"
        sslEnforcement: "Disabled"
        databaseName: "{{ bundle.parameters.database_name }}"
      outputs:
        - name: "MYSQL_URL"
          key: "MYSQL_HOST"
```

In this example, a new output will be created named `MYSQL_URL`. The Azure mixin allows you to specify the key to fetch the output from, in this case it is `MYSQL_HOST`. Each mixin can provide different ways of addressing outputs, so refer to the schema for each mixin. The Porter runtime will keep a map in memory with each of the outputs declared.

TODO: What happens if someone overwrites one? Should we fail the `porter build`?

## Wiring Outputs

Once an output has been declared, it can be referenced in the same way as parameters and credentials. Outputs are referenced with the syntax `"{{ bundle.outputs.OUTPUT_NAME }}"`

For example, given the install step above, we can use the `MYSQL_URL` with the helm mixin in the following way:

```yaml
  - helm:
      description: "Helm Install Wordpress"
      name: porter-ci-wordpress
      chart: stable/wordpress
      set:
        mariadb.enabled: "false"
        externalDatabase.port: 3306
        externalDatabase.host: "{{  bundle.outputs.MYSQL_URL }}"
        externalDatabase.user: "{{ bundle.parameters.mysql_user }}"
        externalDatabase.password: "{{ bundle.parameters.mysql_password }}"
        externalDatabase.database: "{{ bundle.parameters.database_name }}"
```

Just like in the case of credentials and parameters, the value of the `bundle.outputs.MYSQL_URL` reference will be rewritten in the YAML before the helm mixin is invoked.

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
  - helm:
      description: "Helm Install Wordpress"
      name: porter-ci-wordpress
      chart: stable/wordpress
      set:
        image.repository: "{{ bundle.images.ALIAS.repository }}"
        image.tag: "{{ bundle.images.ALIAS.tag }}"
```

## Use Parameters, Credentials, and Outputs from Bundle Dependencies

When using a bundle dependency, you can reference parameters, credentials and outputs in a similar way. To reference things from a dependency, you simply need to use another form of the `"{{  bundle.x.y.z }}"` syntax.

For example, consider a bundle that creates a mysql defined with the following `porter.yaml`:

```yaml
mixins:
- helm

name: mysql
version: 0.1.0
tag: getporter/mysql:v0.1.0

credentials:
- name: kubeconfig
  path: /root/.kube/config

parameters:
- name: database-name
  type: string
  default: mydb
  env: DATABASE_NAME
- name: mysql-user
  type: string
  env: MYSQL_USER

install:
- helm:
    description: "Install MySQL"
    name: porter-ci-mysql
    chart: stable/mysql
    version: 1.6.2
    replace: true
    set:
      mysqlDatabase: "{{ bundle.parameters.database-name }}"
      mysqlUser: "{{ bundle.parameters.mysql-user }}"
    outputs:
    - name: mysql-root-password
      secret: porter-ci-mysql
      key: mysql-root-password
    - name: mysql-password
      secret: porter-ci-mysql
      key: mysql-password
```

In this bundle, we see the normal declaration of credentials, parameters and outputs, along with the use of `"{{  bundle.x.y.z }}"` to use these. With this bundle definition, we can build a second bundle to install wordpress and declare a dependency on this bundle. The `porter.yaml` for this might look something like:

```yaml
mixins:
- helm

name: wordpress
version: 0.1.0
tag: getporter/wordpress:v0.1.0

dependencies:
  mysql:
    tag: getporter/mysql:v0.1.0
    parameters:
      database_name: wordpress
      mysql_user: wordpress

credentials:
- name: kubeconfig
  path: /root/.kube/config

parameters:
- name: wordpress-name
  type: string
  default: porter-ci-wordpress
  env: WORDPRESS_NAME

install:
- helm:
    description: "Install Wordpress"
    name: "{{ bundle.parameters.wordpress-name }}"
    chart: stable/wordpress
    replace: true
    set:
      externalDatabase.database: "{{ bundle.dependencies.mysql.parameters.database-name }}"
      externalDatabase.user: "{{ bundle.dependencies.mysql.parameters.mysql-user }}"
      externalDatabase.password: "{{ bundle.dependencies.mysql.outputs.mysql-password }}"
```

The wordpress bundle declares a dependency on the `mysql` bundle, which we saw above. Now, we are able to refer to the parameters and the outputs from that bundle!

```yaml
install:
- helm:
    description: "Install Wordpress"
    name: "{{ bundle.parameters.wordpress-name }}"
    chart: stable/wordpress
    replace: true
    set:
      externalDatabase.database: "{{ bundle.dependencies.mysql.parameters.database-name }}"
      externalDatabase.user: "{{ bundle.dependencies.mysql.parameters.mysql-user }}"
      externalDatabase.password: "{{ bundle.dependencies.mysql.outputs.mysql-password }}"
```

When the install is executed for this bundle, the steps defined in the `mysql` bundle are completed first. Once those steps have run, any outputs defined are available. In this case, we want to use the `mysql-password` output from the `mysql` dependency. As the example YAML indicates, we do so with the declaration `"{{ bundle.dependencies.mysql.outputs.mysql-password }}"`. The Porter runtime uses the `mysql` manifest to determine how to obtain the values from the dependency output and parameters.

For more information on how dependencies are handled, refer to the [dependencies](/dependencies) documentation.

## Combining References

It is possible to reference multiple parameters, credentials and/or outputs in a single place. You simply combine the expressions as follows:

```yaml
install:
- helm:
    description: "Install Java App"
    name: "{{ bundle.parameters.cool-app}}"
    chart: stable/wordpress
    replace: true
    set:
      jdbc_url: "jdbc:mysql://{{ bundle.outputs.mysql_host }}:{{ bundle.outputs.mysql_port }}/{{ bundle.parameters.database_name }}"
```

[mixin-architecture]: /mixin-dev-guide/architecture/
[how-credentials-work]: /how-credentials-work/
