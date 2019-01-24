---
title: Porter vs. Duffle
description: Does Porter Replace Duffle?
---

<p align="center">NOPE! 😁</p>

## What is Duffle?

[Duffle](https://github.com/deislabs/duffle) is a command-line tool that can be used to build and install [Cloud Native Application Bundles](https://cnab.io) (CNABs). Using Duffle, bundle authors can build, test and distribute installers for their application bundles. The CNAB specification provides high-level guidance for the structure of a bundle, but gives the author a great deal of flexibility. Duffle is a reference implementaiton of the spec and isn't intended to be opinionated. So it also has a great deal of flexibility, at the cost of requiring more knowledge of the CNAB spec when authoring bundles. 

## What is Porter?

Porter is also a command-line tool that can be used to build bundles. It uses a declarative manifest and a special runtime paradigm to make it easier to quickly author bundles, and compose bundles from other porter authored bundles.

## Why Porter?

If Duffle and Porter can both build bundles, why would you use Porter? 

As mentioned above, the CNAB specification defines the general runtime structure of an application bundle. This enables tools like Duffle to support both building and installing bundles. Unfortuanately, with great flexibility comes great complexity. 

Porter on the other hand, provides a declarative authoring experience that encourages the user to adhere to its opionated bundle design. Porter introduces a structured manifest that allows bundle authors to declare dependencies on other bundles, explicitly declare the capabilities that a bundle will use and how parameters, credentials and outputs are passed to individual steps within a bundle. This allows bundle authors to create reusable bundles without requiring extensive knowledge of the CNAB spec.

Porter introduces a command-line tool, along with buildtime and runtime components, called mixins. Mixins give CNAB authors smart componenents that understand how to adapt existing systems, such as Helm, Terraform or Azure, into CNAB actions. After creating a Porter manifest, the command-line tool creates an invocation image and the require CNAB bundle.json. Porter utilizes each mixin to determine the contents of the invocation image. The Porter build command also adds the porter manifest and the porter runtime functionality into the invocation image.  

After a Porter authored bundle is built, it can be run by Duffle or any other CNAB compliant tool. When the bundle is installed, the Porter runtime uses the bundle's porter manifest to determine how to invoke the runtime functionality of each mixin, including how to pass bundle parameters and credentials and how to wire their ouputs to other steps in the bundle. This capability also allows bundle authors to declare dependencies on other Porter bundles and pass the output from one to another.

Porter is **not** intended to replace Duffle, but instead exists to provide an improved bundle authoring experience.

Here is an example that highlights the differences between Porter and Duffle:

## Wordpress Bundle Today

The author has to do everything:
* Create an invocation image with all the necessary binaries and CNAB config.
* Know how to not only install their own application, but how to install/uninstall/upgrade all of their dependencies.
* Figure out CNAB's environment variable naming and how to get at parameters, credentials and actions.

If I write 5 bundles that each use MySQL, I have to redo in each bundle how to manage MySQL. There's no way for someone to write a MySQL bundle that authors can benefit from.

Example:
* [Wordpress Bundle's Dockerfile](https://github.com/deis/bundles/blob/master/wordpress-mysql/cnab/Dockerfile)
* [Wordpress Bundle's Run script](https://github.com/deis/bundles/blob/master/wordpress-mysql/cnab/app/run)

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

# Where's the bundle.json?
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

