<img align="right" src="docs/static/images/porter-logo.png" width="140px" />

[![Build Status](https://dev.azure.com/cnlabs/porter/_apis/build/status/deislabs.porter?branchName=master)](https://dev.azure.com/cnlabs/porter/_build/latest?definitionId=6?branchName=master)


# Porter - We got your baggage, bae

<p align="center"><i>Porter makes authoring bundles easier</i></p>

Porter takes the work out of creating CNAB bundles. It provides a declarative authoring experience that lets you to reuse existing bundles, and understands how to translate CNAB actions to Helm, Terraform, Azure, etc.

# FAQ
* [What is CNAB?](https://cnab.io)
* [Does Porter Replace Duffle?](porter-or-duffle.md)

# Install

## MacOS
```
curl https://deislabs.blob.core.windows.net/porter/latest/install-mac.sh | bash
```

## Linux
```
curl https://deislabs.blob.core.windows.net/porter/latest/install-linux.sh | bash
```

## Windows
```
iwr "https://deislabs.blob.core.windows.net/porter/latest/install-windows.ps1" -UseBasicParsing | iex
```

## Wordpress Bundle Today

The author has to do everything:
* Create an invocation image with all the necessary binaries and CNAB config.
* Know how to not only install their own application, but how to install/uninstall/upgrade all of their dependencies.
* Figure out CNAB's environment variable naming and how to get at parameters, credentials and actions.

If I write 5 bundles that each use MySQL, I have to redo in each bundle how to manage MySQL. There's no way for someone to write a MySQL bundle that authors can benefit from.

Example:
* [Wordpress Bundle's Dockerfile](https://github.com/deis/bundles/blob/master/wordpress-mysql/cnab/Dockerfile)
* [Wordprss Bundle's Run script](https://github.com/deis/bundles/blob/master/wordpress-mysql/cnab/app/run)

## Wordpress Bundle with Porter

CNAB and Duffle provide value to the _consumer_ of the bundle. The bundle development experience still needs improvement. The current state shifts the traditional bash script into a container but doesn't remove the complexity involved in authoring that bash script.

Porter helps you compose bundles with a declarative experience:

* No bash script! 🤩
* No Dockerfile! 😍
* No need to understand the CNAB spec! 😎
* MORE YAML! 🚀

Example:

The porter manifest and runtime handles interpreting and executing the logical package management steps:

```yaml
name: wordpress
version: 0.1.0
invocationImage: deislabs/wordpress:latest

mixins:
  - helm

credentials:
  - name: kubeconfig
    path: /root/.kube/config

parameters:
  - name: wordpress_name
    type: string
    default: mywordpress

install:
  - description: "Install MySQL"
    helm:
      name: mywordpress-mysql
      chart: stable/mysql
      set:
        mysqlDatabase: wordpress
      outputs:
        - name: dbhost
          secret: mywordpress-mysql
          key: mysql-host
        - name: dbuser
          secret: mywordpress-mysql
          key: mysql-user
        - name: dbpassword
          secret: mywordpress-mysql
          key: mysql-password
  - description: "Install Wordpress"
    helm:
      name:
        source: bundle.parameters.wordpress-name
      chart: stable/wordpress
      parameters:
        externalDatabase.database: wordpress
        externalDatabase.host:
          source: bundle.outputs.dbhost
        externalDatabase.user:
          source: bundle.outputs.dbuser
        externalDatabase.password:
          source: bundle.outputs.dbpassword

uninstall:
  - description: "Uninstall Wordpress Helm Chart"
    helm:
      name:
        source: bundle.parameters.wordpress-name
```

## Mixins
Many of the underlying tools that you want to work with already understand package management. Porter makes it easy to
compose your bundle using these existing tools through **mixins**. A mixin handles translating Porter's manifest into
the appropriate actions for the other tools. So far we have mixins for **exec** (bash), **helm** and **azure**.

Anyone can write a mixin binary and drop it into the porter mixins directory (PORTER_HOME/mixins). Mixins are responsible for
a few tasks:

* **Adding lines to the Dockerfile for the invocation image.**

    For example the helm mixin ensures that helm is installed and initialized.
* **Translating steps from the manifest into CNAB actions.**

    For example the helm mixin understands how to install/uninstall a helm chart.
* **Collecting outputs from a step.**

    For example, the step to install mysql handles collecting the database host, username and password.

## Where's the bundle.json?
The `porter build` command handles:

* translating the Porter manifest into a bundle.json
* creating a Dockerfile for the invocation image
* building and pushing the invocation image

So it's still there, but you don't have to mess with it. 😎 A few of the sections in the Porter manifest to map 1:1
to sections in the bundle.json file, such as the bundle metadata, Parameters, and Credentials.

## Hot Wiring a Bundle
I lied, they aren't actually 1:1 mappings. Porter has some special sauce, the `source` reference that makes it much
easier to connect together components in your bundle. For example, creating a database in one step, and then using
the connection string for that database in the next.

Porter supports resolving source values right before a step is executed. Here are a few examples of source references:

* `bundle.outputs.private_key`
* `bundle.parameters.wordpress-name`
* `bundle.credentials.kubeconfig`
* `bundle.dependencies.mysql.parameters.database_name`
* `bundle.dependencies.mysql.outputs.dbhost`

---

## Bundle Dependencies

Porter gets even better when bundles use other bundles. In the example above, the Wordpress installation step relied on the MySQL installation step.
With bundle dependencies, you can rely on other bundles and the outputs that they provide (such as database credentials).

These are _not_ changes to the CNAB runtime spec, though we may later decide that it would be useful to have a companion "authoring" spec. Everything porter does
is baked into your invocation image at build time.

Here's what the example above looks like when instead of shoving everything into a single bundle, we split out installing MySQL from Wordpress into separate bundles.

### MySQL Porter Manifest
The MySQL author indicates that the bundle can provide credentials for connecting to the database that it created.

```yaml
name: mysql
version: 0.1.0
invocationImage: deislabs/mysql:latest

mixins:
  - helm

credentials:
  - name: kubeconfig
    path: /root/.kube/config

parameters:
  - name: database_name
    type: string
    default: mydb

install:
  - description: "Install MySQL"
    helm:
      name: mysql
      chart: stable/mysql
      set:
      mysqlDatabase:
        source: bundle.parameters.database_name
    outputs:
      - name: dbhost
        secret: mysql
        key: mysql-host
      - name: dbuser
        secret: mysql
        key: mysql-user
      - name: dbpassword
        secret: mysql
        key: mysql-password
```

### Wordpress Porter Manifest
```yaml
mixins:
  - helm

name: wordpress
version: 0.1.0
invocationImage: deislabs/wordpress:latest

parameters:
  - name: wordpress_name
    type: string
    default: mywordpress

dependencies:
  - name: mysql
    parameters:
      database_name: wordpress

credentials:
  - name: kubeconfig
    path: /root/.kube/config

install:
  - description: "Install Wordpress"
    helm:
      name:
        source: bundle.parameters.wordpress-name
      chart: stable/wordpress
      set:
        externalDatabase.database:
          source: bundle.dependencies.mysql.parameters.database_name
        externalDatabase.host:
          source: bundle.dependencies.mysql.outputs.dbhost
        externalDatabase.user:
          source: bundle.dependencies.mysql.outputs.dbuser
        externalDatabase.password:
          source: bundle.dependencies.mysql.outputs.dbpassword
```
