---
title: Create a Porter Config File
description: How to customize Porter through configuration file
aliases:
- operators/configuration/
---

Porter's default behavior, such as log level, default plugins, etc., can be modified through its config file.
The config file also provides the ability to define plugin specific
configurations, so you don't need to repeat them each time when you run porter.

## Location

The configuration file should be stored in the PORTER_HOME directory, by default
~/.porter/. In the example below, we are going to use the yaml format, however,
porter supports other file format for its config file as well. For more detailed
information, please see [configuration](/configuration/#config-file).

First check if you have PORTER_HOME set on your machine:
```console
$ echo $PORTER_HOME
```

If it returns nothing, that means your PORTER_HOME is the default value,
~/.porter/. For consistency of this example, let's create the environment
variable like so:
```console
$ echo PORTER_HOME=~/.porter/
```

After making sure you have the PORTER_HOME environment variable set, let's
create the config file in PORTER_HOME.

```console
$ touch $PORTER_HOME/config.yaml
```

## Configure Plugins

By default, Porter uses [host secret plugin](/plugins/host) for resolving secrets and [mongodb-docker](/plugins/mongodb-docker) as the storage
plugin. In the examples below, we will see how to configure Porter to use other plugins. 

### change default secret plugin

The default host plugin does not provide the functionality to work with bundles
that requires/generates sensitive data.
For demoing purpose, we will configure Porter to use the [filesystem](/plugins/filesystem) as the secret plugin so that we can work with sensitive datat.

Once you have the newly created config file open in your choice of editor, let's
change the default secret plugin to [filesystem](/plugins/filesystem):
```yaml
default-secrets-plugin: "filesystem"
```

After saving this change to the config file, you will be able to work with
bundles that contains sensitive parameters or outputs on your local machine. Be
aware that the [filesystem](/plugins/filesystem) does not provide any security
for you data. Please use either the [azure-keyvault](/plugins/azure-keyvault) or [hashicorp-vault](/plugins/hashicorp-vault) plugin in production.

### change default storage plugin

The default mongodb-docker storage plugin runs in a docker container, storing
its data in a separate volume. It's not suitable for production use. 
To switch it out to another storage plugin, like MongoDB, open Porter's config file again and
add below code:
```yaml
# Use the storage configuration named devdb
default-storage = "devdb"
default-storage-plugin: "mongodb"

[[storage]]
  name = "devdb"
  plugin = "mongodb"

  [storage.config]
    url = "conn_str"
    timeout = 10 # time in seconds
```

This plugin allows Porter to store its data in a [MongoDB](/plugins/mongodb) server instance and is
suitable for production use.

If you would like to learn more about Poter's configuration file, see
[here](/configuration).


