---
title: "Upgrade your plugins to securely store sensitive data"
description: "Learn how to keep your sensitive data generated by Porter safe and sound"
date: "2022-05-31"
authorname: "Yingrong Zhao"
author: "@vinozzz"
authorlink: "https://twitter.com/GaysianB612"
authorimage: "https://github.com/VinozzZ.png"
tags: ["best-practices", "plugins"]
summary: |
    Keep sensitive data generated by Porter safe and sound with the new secret plugin protocol
---

As Porter approaches a v1.0.0 release, we have made an improvement in Porter to make sure any sensitive data generated or referenced by Porter is stored in a secure location.
The newly updated secret plugin protocol enables Porter to securely store sensitive data in an external secret store instead of in Porter's database.

Previously Porter only uses plugins for retrieving secrets from a secret store. When it comes to storing data generated by bundles, Porter uses storage plugins like Mongo as its backend database solution. If sensitive data, such as a database connection string, were generated by a bundle, it would be stored in a Mongo database in plain text.
Now Porter requires users to configure a secret store to hold any data that has been marked as sensitive by the bundle. 

Let's walk through how to utilize this new feature by updating your Porter configuration file and selecting an appropriate secret plugin. 

First, [install the latest Porter v1 prerelease](/docs/getting-started/install-porter/#canary).

Next, let's install a bundle that handles sensitive data using just the default Porter configuration.

```
porter install --reference ghcr.io/getporter/examples/sensitive-data --param password=123a123
```

You should see below error message in the output from the above command:
```
failed to save sensitive param to secrete store: rpc error: code = Unknown desc = The default secrets plugin, secrets.porter.host, does not support persisting secrets: not implemented
```

The example bundle defines a sensitive parameter named as `password` and a sensitive output called `name`.

Porter's default secret plugin does not persist sensitive data. Any bundle that references or produces sensitive data will fail to execute. We do this because there isn't a clear set of safe defaults that are suitable for all users when it comes to storing sensitive data. Instead it is up to the user to select and configure an appropriate secrets plugin. 

Now let's configure Porter to persist sensitive data with the [filesystem](/plugins/filesystem/) plugin.

```yaml
default-secrets-plugin: "filesystem"
```

The [filesystem plugin](/plugins/filesystem/) resolves and stores sensitive bundle parameters and outputs as plain-text files in your PORTER_HOME directory.
Note: the filesystem plugin is only intended for testing and local development usage. It's not intended to be used in production. The end of this blog post has recommended plugins that are suitable for production use. 

Now you have a secret store set up, we can finally to install the example bundle, this time successfully.

```
porter install --reference ghcr.io/getporter/examples/sensitive-data --param password=123a123
```

Once the installation process finishes, you should see outputs like below:

```
executing install action from sensitive-data (installation: /sensitive-data)
Install Hello World
Hello, installing example-bundle with password: *******
execution completed successfully!
```

If you inspect Porter's database, it stores a reference to the sensitive data that was saved in the configured secret store. Porter no longer stores the sensitive data in its database.

Instead, we can find our "password" in our filesystem plugin. In your PORTER_HOME directory, you should find a subdirectory named `secrets`. Each file under this directory contains the sensitive value corresponding to a sensitive parameter or sensitive output from a run of a bundle. 

This is why it's important to choose a secure secret plugin for your production environment so that your sensitive data is protected. As you can see, the filesystem plugin is only acceptable for local development and testing.

Here are some secret plugins that we recommend for production use:
- [Azure Key Vault](/plugins/azure/#secrets)
- [Kubernetes Secrets](/plugins/kubernetes/#secrets)
- [Hashicorp Vault](/plugins/hashicorp/)

Give them a try and let us know how it works for you! If there is a secret solution that you would like to use with Porter, let us know, and we can help make that happen more quickly.
