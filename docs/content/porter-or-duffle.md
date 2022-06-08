---
title: Porter or Duffle?
description: A comparison of Porter, Duffle and when you would choose one over the other
---

<h5>A comparison of Porter, Duffle and when you would choose one over the other</h5>

---

* [Does Porter replace Duffle?](#does-porter-replace-duffle)
* [Should I use Porter or Duffle?](#should-i-use-porter-or-duffle)
* [What is Duffle?](#what-is-duffle)
* [What is Porter?](#what-is-porter)
* [Wordpress Bundle with Duffle](#wordpress-bundle-with-duffle)
* [Wordpress Bundle with Porter](#wordpress-bundle-with-porter)

---

## Does Porter replace Duffle?

  <p align="center"><strong>No, Porter is not a replacement of Duffle.</strong></p>

In short:

> Duffle is the reference implementation of the CNAB specification and is used 
> to quickly vet and demonstrate a working specification.

> Porter supports the CNAB spec and empowers bundle authors to create composable, 
> reusable bundles using familiar tools like Helm, Terraform, and their cloud provider's 
> CLIs. Porter is designed to be the best user experience for working with bundles.

## Should I use Porter or Duffle?

If you are contributing to the CNAB specification, we recommend vetting your contributions by
"verification through implementation" on Duffle.

If you are making bundles, may we suggest using Porter?

<p align="center">👩🏽‍✈️ ️️👩🏽‍✈️ 👩🏽‍✈️</p>

## What is Duffle?

[Duffle](https://github.com/cnabio/duffle) is a command-line tool that can be used to build and manage (distribute, install, upgrade and uninstall) [Cloud Native Application Bundles](https://cnab.io) (CNABs). Using Duffle, bundle authors can build, test and distribute installers for their application bundles. The CNAB specification provides high-level guidance for the structure of a bundle, but gives the author a great deal of flexibility. Duffle is the reference implementation of the spec and is not intended to be opinionated. So it also has a great deal of flexibility, at the cost of requiring more knowledge of the CNAB spec when authoring bundles. 

## What is Porter?

Porter is a command-line tool that can also build and manage (distribute, install, upgrade and uninstall) bundles. It uses a declarative manifest and a special runtime paradigm to make it easier to quickly author bundles, and compose bundles from other porter authored bundles.

For some commands, Porter is built _on top of_ the Duffle libraries, using their Go code to implement the
CNAB specification. So there is quite a bit of overlapping functionality.

## Why Porter?

If Duffle and Porter can both build and manage bundles, why would you use Porter? 

As mentioned above, the CNAB specification defines the general structure of an application bundle. This enables tools, like Duffle, to support both building and installing bundles. Unfortunately, with great flexibility comes great complexity. 

Porter on the other hand, provides a declarative authoring experience that encourages the user to adhere to its opinionated bundle design. Porter introduces a structured manifest that allows bundle authors to declare dependencies on other bundles, explicitly declare the capabilities that a bundle will use and how parameters, credentials and outputs are passed to individual steps within a bundle. This allows bundle authors to create reusable bundles without requiring extensive knowledge of the CNAB spec.

Porter introduces a command-line tool, along with buildtime and runtime components, called mixins. Mixins give CNAB authors smart components that understand how to adapt existing systems, such as Helm, Terraform or Azure, into CNAB actions. After creating a Porter manifest, porter creates an invocation image and the required CNAB file structure. Porter utilizes each mixin to determine the contents of the invocation image. The Porter build command also adds the porter manifest and the porter runtime functionality into the invocation image.  

After a Porter-authored bundle is built, it can be run by Duffle or any other CNAB-compliant tool. When the bundle is installed, the Porter runtime uses the bundle's porter manifest to determine how to invoke the runtime functionality of each mixin, including how to pass bundle parameters and credentials and how to wire their outputs to other steps in the bundle. This capability also allows bundle authors to declare dependencies on other Porter bundles and pass the output from one to another.

---

Here is an example that highlights the differences between Porter and Duffle:

## Wordpress Bundle with Duffle

The author has to do everything:

* Create an invocation image with all the necessary binaries and CNAB config.
* Know how to not only install their own application, but how to install/uninstall/upgrade all of their dependencies.
* Figure out CNAB environment variable naming and how to get at parameters, credentials and actions.

If I write 5 bundles that each use MySQL, I have to redo in each bundle how to manage MySQL. There's no way for someone to write a reusable MySQL bundle that authors can benefit from.

Example:

* [Wordpress Bundle's Dockerfile](https://github.com/deis/bundles/blob/master/wordpress-mysql/cnab/Dockerfile)
* [Wordpress Bundle's Run script](https://github.com/deis/bundles/blob/master/wordpress-mysql/cnab/app/run)

CNAB and Duffle provide value to the _consumer_ of the bundle. The bundle development experience still needs improvement. The current state shifts the traditional bash script into a container but doesn't remove the complexity involved in authoring that bash script.

## Wordpress Bundle with Porter

Porter helps you compose bundles with a declarative experience:

* No bash script! 🤩
* No Dockerfile! 😍
* No need to understand the CNAB spec! 😎
* MORE YAML! 🤦‍♀️

Example:

The porter manifest and runtime handles interpreting and executing the logical package management steps:

```yaml
name: wordpress
version: 0.1.0
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
  - name: wordpress_name
    type: string
    default: mywordpress

install:
  - helm3:
      description: "Install MySQL"
      name: mywordpress-mysql
      chart: bitnami/mysql
      set:
        db.name: wordpress
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
  - helm3:
      description: "Install Wordpress"
      name: "{{ bundle.parameters.wordpress-name }}"
      chart: bitnami/wordpress
      version: "9.9.3"
      set:
        externalDatabase.database: wordpress
        externalDatabase.host: "{{ bundle.outputs.dbhost }}"
        externalDatabase.user: "{{ bundle.outputs.dbuser }}"
        externalDatabase.password: "{{ bundle.outputs.dbpassword }}"

uninstall:
  - helm3:
      description: "Uninstall Wordpress Helm Chart"
      releases:
      - "{{ bundle.parameters.wordpress-name }}"
```

### Mixins
Many of the underlying tools that you want to work with already understand package management. Porter makes it easy to
compose your bundle using these existing tools through **mixins**. A mixin handles translating Porter's manifest into
the appropriate actions for the other tools. So far we have mixins for **exec** (bash), **helm**, **azure**, **kubernetes**, and **terraform**.

Anyone can [create a mixin](/mixin-dev-guide). Mixins are responsible for
a few tasks:

* **Adding lines to the Dockerfile for the invocation image.**

    For example the helm mixin ensures that helm is installed and initialized.
* **Translating steps from the manifest into CNAB actions.**

    For example the helm mixin understands how to install/uninstall a helm chart.
* **Collecting outputs from a step.**

    For example, the step to install mysql handles collecting the database host, username and password.

### Where's the bundle.json?
The `porter build` command handles:

* translating the Porter manifest into a bundle.json
* creating a Dockerfile for the invocation image
* building the invocation image

So it's still there, but you don't have to mess with it. 😎 A few of the sections in the Porter manifest to map 1:1
to sections in the bundle.json file, such as the bundle metadata, Parameters, and Credentials.

### Hot Wiring a Bundle
I lied, they aren't actually 1:1 mappings. Porter has templating that makes it
easier to connect together components in your bundle. For example, creating a database in one step, and then using
the connection string for that database in the next.

Porter supports resolving source values right before a step is executed. Here are a few examples of templating:

* `name: "{{ bundle.parameters.wordpress-name }}"`
* `kubeconfig: "{{ bundle.credentials.kubeconfig }}"`
* `externalDatabase.name: "{{ bundle.dependencies.mysql.parameters.database_name }}"`
* `externalDatabase.host: "{{ bundle.dependencies.mysql.outputs.dbhost }}"`

### Bundle Dependencies

Porter gets even better when bundles use other bundles. In the example above, the Wordpress installation step relied on the MySQL installation step.
With bundle dependencies, you can rely on other bundles and the outputs that they provide (such as database credentials).

These are _not_ changes to the CNAB runtime spec, though we may later decide that it would be useful to have a companion "authoring" spec. Everything porter does
is baked into your invocation image at build time.

Here's what the example above looks like when instead of shoving everything into a single bundle, we split out installing MySQL from Wordpress into separate bundles.

### MySQL Porter Manifest
The MySQL author indicates that the bundle can provide credentials for connecting to the database that it created.

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
  - name: database_name
    type: string
    default: mydb

install:
  - helm3:
      description: "Install MySQL"
      name: mysql
      chart: bitnami/mysql
      set:
        db.name: "{{ bundle.parameters.database_name }}"
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

#### Wordpress Porter Manifest
```yaml
mixins:
  - helm3:
      repositories:
        bitnami:
          url: "https://charts.bitnami.com/bitnami"

name: wordpress
version: 0.1.0
registry: getporter

parameters:
  - name: wordpress_name
    type: string
    default: mywordpress

dependencies:
  requires:
    - name: mysql
      bundle:
        reference: getporter/mysql:v0.1.3
      parameters:
        database_name: wordpress

credentials:
  - name: kubeconfig
    path: /home/nonroot/.kube/config

install:
  - helm3:
      description: "Install Wordpress"
      name: "{{ bundle.parameters.wordpress-name }}"
      chart: bitnami/wordpress
      version: "9.9.3"
      set:
        externalDatabase.database: "{{ bundle.dependencies.mysql.parameters.database_name }}"
        externalDatabase.host: "{{ bundle.dependencies.mysql.outputs.dbhost }}"
        externalDatabase.user: "{{ bundle.dependencies.mysql.outputs.dbuser }}"
        externalDatabase.password: "{{ bundle.dependencies.mysql.outputs.dbpassword }}"
```

