---
title: Plugin Types
description: Learn more about available extension points and types of plugins in Porter
weight: 0
---

Porter is extensible and supports a couple extension points where you can alter
its default behavior. Plugins implement a well-defined interface and can be
switched by editing Porter's [configuration file](/configuration/).

## Storage

Storage plugins let you persist files created by Porter to an alternative
location, instead of to the local filesystem under ~/.porter. By default,
credential and parameter sets are saved to the /credentials and /parameters
directories under ~/.porter. Installation records of a bundle being
executed, including claim receipts, action results and outputs, are saved to
the /claims, /results and /outputs directories under ~/.porter.

A storage plugin can implement the [plugins.StorageProtocol interface][storage] and change
where those files are saved. For example, the [Azure plugin](/plugins/azure/)
saves them to Azure Blob Storage.

[storage]: https://github.com/getporter/porter/blob/release/v1/pkg/storage/plugins/storage_protocol.go

## Secrets

Secrets plugins make it easier to securely store and share secret values and
then inject them into a bundle. Currently secrets can be injected as
credentials and parameters. By default, secrets are resolved against the 
localhost: environment variables, files, commands and hard-coded values.

During bundle execution, any parameters or bundle outputs that contains
sensitive data are stored into a secret store that is configured by the user.
Credentials are never persisted, either to Porter's database or secret store,
and are always retrieved just-in-time before the bundle is run.

A secrets plugin can implement the [plugins.SecretsProtocol interface][secretstore] and
resolve credentials from remote and ideally more secure locations. For example,
the [Azure plugin](/plugins/azure/) resolves secrets from Azure Key Vault.

[secretstore]: https://github.com/getporter/porter/blob/release/v1/pkg/secrets/plugins/secrets_protocol.go
