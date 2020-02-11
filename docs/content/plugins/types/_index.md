---
title: Plugin Types
description: Learn more about available extension points and types of plugins in Porter
---

Porter is extensible and supports a couple extension points where you can alter
its default behavior. Plugins implement a well-defined interface and can be
switched by editing Porter's [configuration file](/configuration/).

## Storage

Storage plugins let you persist files created by Porter to an alternative
location, instead of to the local filesystem under ~/.porter. By default,
credential sets (created by `porter credentials generate`) are saved to
~/.porter/credentials and records of a bundle being executed (claims) are saved
to ~/.porter/claims.

A storage plugin can implement the [crud.Store interface][crudstore] and change
where those files are saved. For example, the [Azure plugin](/plugins/azure/)
saves them to Azure Blob Storage.

[crudstore]: https://github.com/cnabio/cnab-go/blob/8ae1722acdeaddc1e720803ca496920c5a4698a2/utils/crud/store.go#L4-L9

## Secrets

Secrets plugins make it easier to securely store and share secret values and
then inject them into a bundle. Currently secrets can only be injected as
credentials but we are working on [injecting them into
parameters](https://github.com/deislabs/porter/issues/878) too. By default,
credentials are resolved against the local host: environment variables, files,
commands and hard-coded values.

A secrets plugin can implement the [secrets.Store interface][secretstore] and
resolve credentials from remote and ideally more secure locations. For example,
the [Azure plugin](/plugins/azure/) resolves secrets from Azure Key Vault.

[secretstore]: https://github.com/cnabio/cnab-go/blob/8ae1722acdeaddc1e720803ca496920c5a4698a2/secrets/store.go#L4-L13
