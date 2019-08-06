---
title: gcloud mixin
description: Using the gcloud mixin
---

Run a command using the [gcloud CLI](https://cloud.google.com/sdk/gcloud/reference/).

Source: https://github.com/deislabs/porter-gcloud

### Install or Upgrade
```
porter mixin install gcloud --feed-url https://cdn.deislabs.io/porter/atom.xml
```

## Mixin Syntax

See the [gcloud CLI Command Reference](https://cloud.google.com/sdk/gcloud/reference/) for the supported commands

```yaml
gcloud:
  description: "Description of the command"
  groups: GROUP
  command: COMMAND
  arguments:
  - arg1
  - arg2
  flags:
    FLAGNAME: FLAGVALUE
    REPEATED_FLAG:
    - FLAGVALUE1
    - FLAGVALUE2
```

```yaml
gcloud:
  description: "Description of the command"
  groups:
  - GROUP 1
  - GROUP 2
  command: COMMAND
  arguments:
  - arg1
  - arg2
  flags:
    FLAGNAME: FLAGVALUE
    REPEATED_FLAG:
    - FLAGVALUE1
    - FLAGVALUE2
```

## Examples

The [Compute Example](https://github.com/deislabs/porter-gcloud/tree/master/examples/compute) provides a full working bundle demonstrating how to use this mixin.

### Authenticate

```yaml
gcloud:
  description: "Authenticate"
  groups:
    - auth
  command: activate-service-account
  flags:
    key-file: gcloud.json
```

### Provision a VM

```yaml
gcloud:
  description: "Create VM"
  groups:
    - compute
    - instances
  command: create
  arguments:
    - porter-test
  flags:
    project: porterci
    zone: us-central1-a
    machine-type: f1-micro
    image: debian-9-stretch-v20190729
    image-project: debian-cloud
    boot-disk-size: 10GB
    boot-disk-type: pd-standard
    boot-disk-device-name: porter-test
```
