---
title: Using Parameters, Credentials and Outputs
description: How to wire parameters, credentials and outputs into steps
---

# Parameters, Credentials and Outputs in Porter

In the Porter manifest, you can declare both [parameters](https://github.com/deislabs/cnab-spec/blob/master/101-bundle-json.md#parameters) and [credentials](https://github.com/deislabs/cnab-spec/blob/master/101-bundle-json.md#credentials), which are defined in the CNAB spec. The CNAB specification specifies how both credentials and parameters can be provided to an [invocation image](https://github.com/deislabs/cnab-spec/blob/master/102-invocation-image.md) and how they should be declared in a bundle manifest. 

The Porter manifest allows you to provide all the elements defined in the specification, but with Porter's opinionated approach by default. In addition to providing a mechanism for declaring parameters and credentials at the bundle level, Porter provides a way to declare how each of these are provided to [mixins](/mixin-architecture). This mechanism is also applicable to declaring how output from one mixin can be passed to another, as well as how to consume parameters, credentials and outputs from bundle dependencies.

## Parameters

In order to declare a parameter in a Porter bundle, you first declare a parameters block with one or more parameters in the `porter.yaml`. For example, to declare a parameter named database_name, you might include the following YAML block:

```yaml
parameters:
- name: database_name
  type: string
```

This is the minimum required to create a parameter in Porter. Porter will specify an environment variable destination that defaults to the upper-cased name of the parameter. This would result in the following parameter definition in the `bundle.json`:

```json
"parameters": {
    "database_name": {
        "type": "string",
        "required": false,
        "metadata": {},
        "destination": {
            "path": "",
            "env": "DATABASE_NAME"
        }
    }
}
```

You can also provide any other attributes, as specified by the CNAB [parameters](https://github.com/deislabs/cnab-spec/blob/master/101-bundle-json.md#parameters) specification. To specify a default value, for example, you could provide the following parameter definition:

```yaml
- name: database_name
  type: string
  default: "wordpress"
```

This would result in the following `bundle.json` parameter definition:

```json
"parameters": {
    "database_name": {
        "type": "string",
        "defaultValue": "wordpress",
        "required": false,
        "metadata": {},
        "destination": {
            "path": "",
            "env": "DATABASE_NAME"
        }
    }
}
```

## Wiring Parameters

Once a parameter has been declared in the `porter.yaml`, Porter provides a simple mechanism for specifying how the parameter value should be wired into the mixin. To reference a parameter in the bundle, you use a notation of the form `source: bundle.parameters.PARAM_NAME`. This can be used anywhere within step definition. For example, to use the `database_name` parameter defined above in the `set` block of the helm mixin:

```yaml
install:
- description: "Install MySQL"
  helm:
    name: porter-ci-mysql
    chart: stable/mysql
    version: 0.10.2
    replace: true
    set:
      mysqlDatabase:
        source: bundle.parameters.database-name
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
      command:
        source:  bundle.parameters.command
```

This syntax is used in dictionary, as above, or in an list:

```yaml
parameters:
  - name: command
    type: string
    default: "echo Hello World"

install:
  - description: "Install Hello World"
    exec:
      command: bash
      arguments:
        - -c
        - source:  bundle.parameters.command
```

When the bundle is executed, the Porter runtime will locate the parameter definition in the `porter.yaml` to determine where the parameter value has been stored. The Porter runtime will then rewrite the YAML block before it is passed to the mixin. For example, given the YAML example above, the exec mixin will actually get the following YAML:

```yaml
description: "Install Hello World"
exec:
  command: bash
  arguments:
    - -c
    - echo Hello World
```

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

The same mechanism for declaring how to use a parameter can be used for credentials. To declare a credential usage, references are defined with the following syntax: `source: bundle.credentials.CREDENTIAL_NAME`.

When the bundle is executed, the Porter runtime will locate the parameter definition in the `porter.yaml` to determine where the parameter value has been stored. The Porter runtime will then rewrite the YAML block before it is passed to the mixin.

## Outputs

In addition to parameters and credentials, Porter allows you to declare how to store mixin outputs. The form used to declare these varies by mixin, but might look something like:

```yaml
install:
  - description: "Create Azure MySQL"
    azure:
      type: mysql
      name: demo-mysql-azure-porter-demo-wordpress
      resourceGroup: "porter-test"
      parameters:
        administratorLogin:
          source: bundle.parameters.mysql_user
        administratorLoginPassword:
          source: bundle.parameters.mysql_password
        location: "eastus"
        serverName: "mysql-jeremy-porter-test-jan-2018"
        version: "5.7"
        sslEnforcement: "Disabled"
        databaseName:
          source: bundle.parameters.database_name
    outputs:
      - name: "MYSQL_URL"
        key: "MYSQL_HOST"
```

In this example, a new output will be created named `MYSQL_URL`. The Azure mixin allows you to specity the key to fetch the output from, in this case it is `MYSQL_HOST`. Each mixin can provide different ways of addressing outputs, so refer to the schema for each mixin. The Porter runtime will keep a map in memory with each of the outputs declared.

TODO: What happens if someone overwrites one? Should we fail the `porter build`?

## Wiring Outputs

Once an output has been declared, it can be referenced in the same way as parameters and credentials. Outputs are referenced with the syntax `source: bundle.outputs.OUTPUT_NAME`

For example, given the install step above, we can use the `MYSQL_URL` with the helm mixin in the following way:

```yaml
  - description: "Helm Install Wordpress"
    helm:
      name: porter-ci-wordpress
      chart: stable/wordpress
      set:
        mariadb.enabled: "false"
        externalDatabase.port: 3306
        externalDatabase.host:
          source: bundle.outputs.MYSQL_URL
        externalDatabase.user:
          source: bundle.parameters.mysql_user
        externalDatabase.password:
          source: bundle.parameters.mysql_password
        externalDatabase.database:
          source: bundle.parameters.database_name
```

Just like in the case of credentials and parameters, the value of the `bundle.outputs.MYSQL_URL` reference will be rewritten in the YAML before the helm mixin is invoked.

## Using Parameters, Credentials, and Outputs from Bundle Dependencies

When using a bundle dependency, you can reference parameters, credentials and outputs in a similar way. To reference things from a dependency, you simply need to use another form of the `source: ....` syntax.

For example, consider a bundle that creates a mysql defined with the following `porter.yaml`:

```yaml
mixins:
- helm

name: mysql
version: 0.1.0
invocationImage: porter-mysql:latest

credentials:
- name: kubeconfig
  path: /root/.kube/config

parameters:
- name: database-name
  type: string
  default: mydb
  destination:
    env: DATABASE_NAME
- name: mysql-user
  type: string
  destination:
    env: MYSQL_USER

install:
- description: "Install MySQL"
  helm:
    name: porter-ci-mysql
    chart: stable/mysql
    version: 0.10.2
    replace: true
    set:
      mysqlDatabase:
        source: bundle.parameters.database-name
      mysqlUser:
        source: bundle.parameters.mysql-user
  outputs:
  - name: mysql-root-password
    secret: porter-ci-mysql
    key: mysql-root-password
  - name: mysql-password
    secret: porter-ci-mysql
    key: mysql-password
```

In this bundle, we see the normal declaration of credentials, parameters and outputs, along with the use of `source: ....` to use these. With this bundle definition, we can build a second bundle to install wordpress and declare a dependency on this bundle. The `porter.yaml` for this might look something like:

```yaml
mixins:
- helm

name: wordpress
version: 0.1.0
invocationImage: porter-wordpress:latest

dependencies:
- name: mysql
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
  destination:
    env: WORDPRESS_NAME

install:
- description: "Install Wordpress"
  helm:
    name:
      source: bundle.parameters.wordpress-name
    chart: stable/wordpress
    replace: true
    set:
      externalDatabase.database:
        source: bundle.dependencies.mysql.parameters.database-name
      externalDatabase.user:
        source: bundle.dependencies.mysql.parameters.mysql-user
      externalDatabase.password:
        source: bundle.dependencies.mysql.outputs.mysql-password
```

The wordpress bundle declares a dependency on the `mysql` bundle, which we saw above. Now, we are able to refer to the parameters and the outputs from that bundle!

```yaml
install:
- description: "Install Wordpress"
  helm:
    name:
      source: bundle.parameters.wordpress-name
    chart: stable/wordpress
    replace: true
    set:
      externalDatabase.database:
        source: bundle.dependencies.mysql.parameters.database-name
      externalDatabase.user:
        source: bundle.dependencies.mysql.parameters.mysql-user
      externalDatabase.password:
        source: bundle.dependencies.mysql.outputs.mysql-password
```

When the install is executed for this bundle, the steps defined in the `mysql` bundle are completed first. Once those steps have run, any outputs defined are available. In this case, we want to use the `mysql-password` output from the `mysql` dependency. As the example YAML indicates, we do so with the declaration `source: bundle.dependencies.mysql.outputs.mysql-password`. The Porter runtime uses the `mysql` manifest to determine how to obtain the values from the dependency output and parameters.

For more information on how dependencies are handled, refer to the [dependencies](/dependencies) documentation.