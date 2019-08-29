---
title: gcloud mixin
description: Using the gcloud mixin
---

<img src="/images/mixins/google.png" class="mixin-logo" />

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
  outputs:
    - name: NAME
      jsonPath: JSONPATH
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
  outputs:
    - name: NAME
      jsonPath: JSONPATH
```


### Outputs

The mixin supports `jsonpath` outputs.


#### JSON Path

The `jsonPath` output treats stdout like a json document and applies the expression, saving the result to the output.

```yaml
outputs:
- name: NAME
  jsonPath: JSONPATH
```

For example, if the `jsonPath` expression was `$[*].id` and the command sent the following to stdout: 

```json
[
  {
    "id": "1085517466897181794",
    "name": "my-vm"
  }
]
```

Then then output would have the following contents:

```json
["1085517466897181794"]
```

---

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
  outputs:
    - name: vms
      jsonPath: "$[*].id"
```
