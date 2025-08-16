---
title: Host Secrets Plugin
description: Resolve secrets using the local host
---

The Host Secrets plugin is built-in to Porter. It resolves secrets referenced in
a parameter or credential set using values available within the host environment
such as: environment variables, files, command output, and hard-coded values.
This plugin is suitable for development and test but is not recommended for
production use. In production, we recommend using a plugin that integrates with
a remote secret store, such as the [Azure Key Vault] or [Hashicorp Vault]
plugins.

[Azure Key Vault]: /plugins/azure/#secrets
[Hashicorp Vault]: /plugins/hashicorp/

## Plugin Configuration

There is no configuration available for the host plugin.