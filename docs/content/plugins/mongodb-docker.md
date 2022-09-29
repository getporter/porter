---
title: MongoDB Docker Storage Plugin
description: A built-in plugin that stores Porter's data in a container running MongoDB.
---

The MongoDB Docker Storage plugin is built-in to Porter and is the default
storage plugin. This plugin is suitable for development and test but should not
be used in production.

The plugin runs a MongoDB server in a container, storing its data on a separate
volume. The container is named `porter-mongodb-docker-plugin` and the volume is
named `porter-mongodb-docker-plugin-data`. The plugin leaves the container
running in-between Porter commands for performance reasons. It is safe to stop
or remove the container. Removing the volume will result in data loss.

## Plugin Configuration

No configuration is required to use the default storage plugin. However, you may
configure the port if there is a conflict with the default port, 27018.

```toml
default-storage = "mymongo"

[[storage]]
  name = "mymongo"
  plugin = "mongodb-docker"

  [storage.config]
    port = "27019"
```

[config file]: /configuration/#config-file

## Config Parameters

### port

The port parameter configures which port the MongoDB server listens on. By default, this plugin listens on 27018.


## Remove Plugin Data

If you want to do a fresh installation of Porter and start over with a new database, run the following commands to remove the container and volume used by the plugin.

```
docker rm -f porter-mongodb-docker-plugin
docker volume rm porter-mongodb-docker-plugin-data
```
