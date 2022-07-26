---
title: "Get those secrets out of your config"
description: "Learn how to keep your Porter config file secret-free"
date: "2022-02-01"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://twitter.com/carolynvs"
authorimage: "https://github.com/carolynvs.png"
tags: ["best-practices"]
summary: |
    Keep sensitive data out of your Porter config files with the new variables ${env.NAME} and ${secret.KEY}
---

As part of the v1 hardening process, we have been hard at work securing Porter.
Keeping sensitive data off your machine, and safely in a secret store or vault where it belongs is a big part of that.
Now Porter's config file is getting the same white glove treatment that bundle credentials always have!
We have [added templating support to Porter's config file][cfg-docs] so that you can use environment variables and secrets without hard-coding sensitive data in the file.

[cfg-docs]: /configuration/#config-file

Porter has plugins for retrieving secrets from a secret store, and for storing its data in a Mongo database.
Configuring the plugins with credentials to connect to those resources is how sensitive data sneaks into the config file.
Let's walk through what your config file may look like today and how to take advantage of templating to keep your config file secret-free.

```toml
# ~/.porter/config.toml
default-storage = "mydb"

[[storage]]
  name = "mydb"
  plugin = "mongodb"

  [[storage.config]]
    url = "a top secret mongodb connection string"

[[secrets]]
  name = "mysecrets"
  plugin = "azure.keyvault"
  
  [[secrets.config]]
    vault = "myvault"
    subscription-id = "my azure subscription id"
```

In the example above we have two bits of sensitive data in the config file: a mongodb connection string, and our Azure subscription id.
We can replace those hard-coded values with templates, using `${env.NAME}` to insert an environment variable and `${secret.KEY}` to resolve a secret from your default secret store. Below is the final, secret-free version of the same config file:

```toml
# ~/.porter/config.toml
default-storage = "mydb"

[[storage]]
  name = "mydb"
  plugin = "mongodb"

  [[storage.config]]
    url = "${secret.porter-connection-string}"

[[secrets]]
  name = "mysecrets"
  plugin = "azure.keyvault"
  
  [[secrets.config]]
    vault = "myvault"
    subscription-id = "${env.AZURE_SUBSCRIPTION_ID}"
```

Porter takes two passes over the configuration file: first replacing environment variables, and then resolving any secrets used in the config file with the secrets plugin.

You may have noticed that this is a different template syntax than what is used in porter.yaml.
In an upcoming release, Porter's template syntax will get a similar refresh, so that the templating language is the same for both Porter's manifest and its configuration file.

Install the [latest v1 prerelease](https://github.com/getporter/porter/releases?q=v1.0.0) and make your config files secret-free! 🕊
