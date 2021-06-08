---
title: Hashicorp plugin
description: Integrate Porter with Hashicorp
---

<img src="/images/plugins/hashicorp.png" class="mixin-logo" style="width: 300px"/>

Integrate Porter with Hashicorp Vault.

Source: https://github.com/dev-drprasad/porter-hashicorp-plugins

## Install or Upgrade

**Note:** Supports porter version greater or equal to `v0.23.0-beta.1` and supports only `KV Version 2 secret engine`.

```
porter plugin install hashicorp --feed-url https://github.com/dev-drprasad/porter-hashicorp-plugins/releases/download/feed/atom.xml
```

## Plugin Configuration

To use vault plugin, add the following config to porter's config file (default location: `~/.porter/config.toml`). Replace `vault_addr`, `vault_token` and `path_prefix` with proper values.

```toml
default-secrets = "porter-secrets"
[[secrets]]
name = "porter-secrets"
plugin = "hashicorp.vault"

[secrets.config]
vault_addr = "http://vault.example.com:7500"
path_prefix = "organization/team/project"
vault_token = "token"
```

## Config Parameters

### path_prefix

`path_prefix` lets allow you to specify prefix for your secret path. Let' say you have a secret (`myawesomeproject`) with path `organization/team/project/myawesomeproject`, then you can configure `path_prefix` as `organization/team/project`.