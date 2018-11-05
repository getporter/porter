# Porter - We got your baggage, bae

<p><center><i>Porter makes authoring bundles easier</i></center></p>

Porter is a helper binary that Duffle can build into your CNAB invocation images. It provides a declarative authoring experience for CNAB bundles that allows you to reuse existing bundles, and understands how to translate CNAB actions to Helm, Terraform, Azure, etc.

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

Porter replaces the bash script with a declarative experience:

* No run script! 🤩
* No Dockerfile! 😍
* No need to understand the CNAB spec! 😎
* MORE YAML! 🚀

Example:

The porter.yaml file replaces the bundle run script. The run script is now just a boring call to `porter porter.yaml`. The porter runtime handles interpreting and executing the steps:

```yaml
mixins:
  - helm
dependencies:
  - name: mysql
    parameters:
      database-name: wordpress
    outputs:
      - source: bundle.dependencies.mysql.outputs.host
        destination: bundle.credentials.dbhost
install:
  - name: "Install Wordpress Helm Chart"
    helm: 
      name:
        source: bundle.name
      chart: stable/wordpress
      parameters:
        externalDatabase.database:
          source: bundle.dependencies.mysql.parameters.database-name
        externalDatabase.host:
          source: bundle.credentials.dbhost
        externalDatabase.user:
          source: bundle.credentials.dbuser
        externalDatabase.password:
          source: bundle.credentials.dbpassword
uninstall:
  - name: "Uninstall Wordpress Helm Chart"
    helm:
      name:
        source: bundle.name
```

* The porter binary calls the helm mixin which handles running helm init and helm install using all the values above
* Porter handles populating Parameters from the `CNAB_P_*` environment variables
* Porter makes the credential environment variables easyily accessible
---

## Bundle Dependencies

Porter gets even better when bundles can use other bundles. In the Wordpress example above, the Wordpress bundle was able to access database connection variables provided by the MySQL bundle.

These are _not_ changes to the CNAB runtime spec, though we may later decide that it would be useful to have a companion "authoring" spec. Everything porter does
is baked into your invocation image at build time.

## MySQL porter.yaml
The MySQL author indicates that the bundle can provide credentials for connecting to the database that it created.

```yaml
mixins:
  - azure

outputs:
  - name: host
    env: MYSQL_HOST
  - name: user
    env: MYSQL_USER

install:
  - name: "Provision MySQL Server"
    azure:
      mysql.server: # The mixin exports MYSQL_HOST, MYSQL_USER, etc
       - create:
           resource-group: default
           location: westus
           sku-name: GP_Gen4_2
           version: "5.7"
  - name: "Create MySQL Database"
    azure:
      mysql.db:
        - create:
          name:
            source: bundle.parameters.server-name
          server-name:
            source: bundle.outputs.host
          resource-group: default
```

## Piping dependency inputs and outputs
The Wordpress author can connect the credentials provided by the MySQL bundle directly to the Wordpress database credentials that the Helm Wordpress chart requires.

```yaml
...
dependencies:
  - name: mysql
    parameters:
      database-name: wordpress
    connections:
      - source: dependencies.mysql.outputs.host
        destination: credentials.dbhost
...
```
