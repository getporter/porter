---
title: Data Migration
description: How to prepare for and migrate Porter's data
---

Hello friend, if you have found this page then Porter's storage format has
changed. You need to back up Porter's data and then migrate it to the current
version. I'm sorry we dumped this on you and we'll try to limit how often this
happens in the future!

Until the data is migrated, newer versions of Porter will halt and request a 
migration. The migration is one-way, you may continue to use an [older version
of Porter][install-old] to delay migrating your data until a more appropriate time.

1. [Backup](#backup)
2. [Migrate](#migrate)

[install-old]: /install/#older-version

## Backup

Backup Porter's data before performing the migration. You should always backup
the Porter home directory, usually **~/.porter**.

You may have data stored in an additional remote location depending on the
plugin that you are using. Open up your Porter config file located at
~/.porter/config.toml.

* If you don't have one, congratulations!, you are using the filesystem plugin 
  and all of your data is in Porter home.
* If you do have a config file, take a look and determine if you are using a 
  different storage plugin.
  
NOTE: Only storage plugins have data that requires migration, not secret plugins.

### Azure Plugin

The Azure plugin stores files in Azure Blob Storage. You should also backup
the container named "porter" in the storage account in addition to backing
up Porter home on your local file system.

Versions v0.9.0+ of the Azure plugin requires that the blob storage account
has the Blob Index preview enabled on the account. [Only certain regions support
the Blob Index preview, and you must run a command to enable it on your
account][blob-preview].

At the time of the v0.28.0 Porter release, only the following Azure regions
support the Blob Index preview. If your current storage account is
in another region, you can create a new account in one of the regions 
below and copy the container to the new account.

* Canada Central
* Canada East
* France Central
* France South

The following az cli command enables the Blob Index preview on your account (it
can take up to 10 minutes for this command to propagate):

```
az feature register --namespace Microsoft.Storage --name BlobIndex
az provider register --namespace 'Microsoft.Storage'
```

After you have created a new account an appropriate region and enabled the
Blob Index preview, you can use azcopy to [copy the porter container to a new
storage account][azcopy-container].

[blob-preview]: https://docs.microsoft.com/en-us/azure/storage/blobs/storage-manage-find-blobs?tabs=azure-portal#regional-availability-and-storage-account-support
[azcopy-container]: https://docs.microsoft.com/en-us/azure/storage/common/storage-use-azcopy-blobs#copy-a-container-to-another-storage-account
 
## Migrate

Once you have completed a backup of Porter's data, you are ready to run the migration
by running the following command:

```
porter storage migrate
```

After the migration completes, Porter records the new storage version in
~/.porter/schema.json, and you can use all of Porter's commands again.