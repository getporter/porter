---
title: Filesystem Secrets Plugin
description: Resolve secrets using the local filesystem
---

The Filesystem secrets plugin is an internal plugin that can be enabled through Porter's configuration file.
It stores and resolves sensitive bundle parameters and outputs as plaintext files in your PORTER_HOME directory.
This plugin is suitable for development and test but is not recommended for production use.
In production, we recommend using a plugin that integrates with a remote secret store, such as the [Azure Key Vault] or [Hashicorp Vault]
plugins.

[Azure Key Vault]: /plugins/azure/#secrets
[Hashicorp Vault]: /plugins/hashicorp/

## Plugin Configuration

There is no configuration available for the filesystem plugin.

