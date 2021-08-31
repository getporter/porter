---
title: MongoDB Storage Plugin
description: A built-in plugin that stores Porter's data in MongoDB.
---

The MonggoDB storage plugin is built-in to Porter. The plugin allows Porter to
store its data in a MongoDB server. This plugin is suitable for production use.

## Plugin Configuration

To use the mongodb plugin, add the following config to porter's [config file]. Replace `conn_str` with the
connection string for your MongoDB server.

```toml
default-secrets = "mymongo"

[[storage]]
  name = "mymongo"
  plugin = "mongodb"

  [storage.config]
    url = "conn_str"
    timeout = 10 # time in seconds
```

[config file]: /configuration/#config-file

## Config Parameters

### url

The url configuration parameter specifies how to connect to a MongoDB server.
The general format is below. See the [MongoDB Connection
String](https://docs.mongodb.com/manual/reference/connection-string/)
documentation for more details.

```
"mongodb://USER:PASSWORD@HOST:PORT/DATABASE/?OPTIONS"
```

Only the host portion is required. The port defaults to "27017", and the
database to "porter". Porter will create the database, collections and indices
if they do not already exist.

Here is an example connection string for an instance of MongoDB running on
localhost at the default port 27017, using the database name "mydb".

```
"mongodb://localhost:27017/mydb"
```

### timeout

Sets the timeout (in seconds) used for database queries.
The default timeout is 2 seconds.
