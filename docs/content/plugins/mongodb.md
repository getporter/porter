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
default-storage = "mymongo"

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
The default timeout is 10 seconds.

## Cosmos DB

You can use Azure Cosmos DB API for MongoDB as a database with the mongodb plugin.
The connection string to the database must be modified to work with Porter, the only supported query string option is `ssl=true`.
See the [Azure documentation for retrieving the ComosDB connection string](https://docs.microsoft.com/en-us/azure/cosmos-db/mongodb/connect-mongodb-account#get-the-mongodb-connection-string-to-customize).
For example, if your connection string looks like this:

```
mongodb://mydb:mykey@mydb.mongo.cosmos.azure.com:10255/?ssl=true&replicaSet=globaldb&retrywrites=false&maxIdleTimeMS=120000&appName=@mydb@
```

You need to remove the additional options and leave only the `ssl=true` option like so:

```
mongodb://mydb:mykey@mydb.mongo.cosmos.azure.com:10255/?ssl=true&
```

If you forget to remove the additional options from the connection string, you may see an error like the one below when connecting:

```
server selection error: context deadline exceeded, current topology: { Type: Unknown, Servers: [{ Addr: mydb.mongo.cosmos.azure.com:10255, Type: Unknown }, ] }
```

⚠️ There is a known performance issue when using CosmosDB with Porter that needs to be addressed: https://github.com/getporter/porter/issues/1782
