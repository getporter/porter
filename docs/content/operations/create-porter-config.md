---
title: Create a Porter Config File
description: Learn how to customize Porter's default behavior through configuration file
---

Porter's default behavior, such as log level and default plugins, can be modified through its config file.
The config file also provides the ability to define plugin specific
configurations, so you don't need to repeat them each time when you run porter.

## Create a config file

The configuration file should be stored in the PORTER_HOME directory, by default
~/.porter/. In the example below, we are going to use the yaml format, however,
porter supports other file formats for its config file as well. For more detailed
information, please see [configuration](/configuration/#config-file).

Create a file at ~/.porter/config.yaml and open it in your editor.

## Configure Plugins

By default, Porter uses the [host plugin](/plugins/host) for resolving secrets and the [mongodb-docker](/plugins/mongodb-docker) as the storage
plugin. In the examples below, we will see how to configure Porter to use other plugins. 

### Change the default secrets plugin

The default host plugin does not provide the functionality to work with bundles
that work with sensitive data.
Next, configure Porter to use the [filesystem](/plugins/filesystem) as the secret plugin so that you can work with sensitive data.
The filesystem plugin stores secrets in your PORTER_HOME directory. It is suitable for trying out Porter, and local development and testing, but should not be used in production.

Once you have the [newly created config file](#create-a-config-file), open in your choice of editor,
change the default secret plugin to [filesystem](/plugins/filesystem):
```yaml
default-secrets-plugin: "filesystem"
```

After saving this change to the config file, you will be able to work with
bundles that contains sensitive parameters or outputs on your local machine. Be
aware that the [filesystem](/plugins/filesystem) stores sensitive data in plaintext on your filesystem.
Use either the [azure-keyvault](/plugins/azure-keyvault) or [hashicorp-vault](/plugins/hashicorp-vault) plugin in production.

### Change the default storage plugin

The default mongodb-docker storage plugin runs in a docker container, storing
its data in a volume named `porter-mongodb-docker-plugin-data`. It's not suitable for production use. 
To switch it out to another storage plugin, like MongoDB, open Porter's config file again and
add below code, setting the url property to a valid connection string:
```yaml
# Use the storage configuration named devdb
default-storage: "devdb"

storage:
  name: "devdb"
  plugin: "mongodb"
  config:
    # Set the url property to a valid Mongo connection string
    url: "TODO_REPLACE"
```

The [mongodb plugin](/plugins/mongodb/) allows Porter to store its data in a MongoDB server instance and is
suitable for production use.

## Next Steps
* [Configuration File Format](/configuration/)
* [Available Plugins](/plugins/#available-plugins)

