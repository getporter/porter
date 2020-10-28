---
title: Kubernetes Mixin 
description: The Porter Kubernetes Mixin 
---

# Kubernetes Mixin

## Overview

The Kubernetes Mixin provides bundle authors with the ability to apply and delete Kubernetes manifests. The mixin will leverage `kubectl`, similar to how the Helm mixin utilizes the `helm` command line tool.

## Authoring a bundle with the mixin

In order to build a CNAB with the kubernetes mixin, the bundle author should place one more more Kubernetes manifests within the cnab directory of their bundle. The mixin will provide a default location, `/cnab/app/kubernetes`, but bundle authors can place manifests in any path within the cnab directory and then declare the directory within the `porter.yaml`. This will enable the mixin to be used multiple times within a bundle with different manifests.

## Buildtime

Rather than trying to rebuild the functionality of `kubectl`, this mixin will contribute lines to the invocation image Dockerfile that will result in `kubectl` being installed:

```
RUN apt-get update && \
apt-get install -y apt-transport-https curl && \
curl -o kubectl https://storage.googleapis.com/kubernetes-release/release/v1.13
.0/bin/linux/amd64/kubectl && \
mv kubectl /usr/local/bin && \
chmod a+x /usr/local/bin/kubectl
```

Releases of the mixin will pin to versions of `kubectl`. We will initially pin to the 1.13 release, but may bump to 1.14 depending on implementation schedule and the kubernetes release cycle.

## Dry Run

The mixin will use `kubectl apply --dry-run` in order to perform a dry run for the bundle.

## Run Time

### Credentials

The Kubernetes Mixin requires a kubeconfig file. The mixin will allow the user to specify where it is mounted at, but will assume it is provided at `/root/.kube/config` if not otherwise specified. The kubeconfig should have sufficient privileges to apply the resources included in the bundle.

### Install

At runtime, the mixin will use the `kubectl apply` command when an `install` action is specified. This will result in the resources defined in the supplied manifests being created, as they should not currently exist. The use of the `apply` command will allow the use of the `wait` flag The mixin will not support all of the [options](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply) available, specifically `dry-run`, or options related to editing or deleting resources. Available parameters are spelled out below.

#### Parameters

The mixin allows bundle authors to specify the following parameters on install:

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `namespace` | string | The namespace in which to create resources | `default` |
| `manifests` | string | The path to the manifests. Can be a file or directory | `/cnab/app/kubernetes` |
| `record` | boolean | Record current kubectl command in the resource annotation. If set to false, do not record the command. If set to true, record the command. If not set, default to updating the existing annotation value only if one already exists. | `false` | 
| `selector` | string | Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2) | |
| `validate` | boolean | If true, use a schema to validate the input before sending it | `true` |
| `wait` | boolean | If true, wait for resources to be gone before returning. This waits for finalizers. | `true` |

### Upgrade

 At runtime, the mixin will use the `kubectl apply` command when an `upgrade` action is specified. This will result in the resources defined in the supplied manifests being created or deleted, as appropriate. As the manifests will be contained within the bundle's invocation image, an upgrade action against an invocation image that was used for install is a no-op. The use of the `apply` command will allow the use of the `wait` flag The mixin will not support all of the [options](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply) available, specifically `dry-run`, or options related to editing or deleting resources. Available parameters are spelled out below.

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `namespace` | string | The namespace in which to create resources. | `default` |
| `manifests` | string | The path to the manifests. Can be a file or directory. | `/cnab/app/kubernetes` |
| `force` | boolean | If true, immediately remove resources from API and bypass graceful deletion. Note that immediate deletion of some resources may result in inconsistency or data loss and requires confirmation. Overrides `gracePeriod`. | `false`|
| `gracePeriod` | integer | Period of time in seconds given to the resource to terminate gracefully. Ignored if negative. Set to 1 for immediate shutdown. If `force` is true, will result in 0. | -1 |
| `overwrite` | boolean | Automatically resolve conflicts between the modified and live configuration by using values from the modified configuration. | `true` |
| `prune` | boolean | Automatically delete resource objects, including the uninitialized ones, that do not appear in the configs. | `false` |
| `record` | boolean | Record current kubectl command in the resource annotation. If set to false, do not record the command. If set to true, record the command. If not set, default to updating the existing annotation value only if one already exists. | `false` ||
| `selector` | string | Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2). | |
| `timeout` | integer | The length of time (in seconds) to wait before giving up on a delete, zero means determine a timeout from the size of the object. | 0 |
| `validate` | boolean | If true, use a schema to validate the input before sending it. | `true` |
| `wait` | boolean | If true, wait for resources to be gone before returning. This waits for finalizers. | `true` |

### Delete

At runtime, the mixin will use the `kubectl delete` command to remove the resources specified by the bundle manifests.

#### Parameters

The mixin allows bundle authors to specify the following parameters on delete:

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `namespace` | string | The namespace in which to create resources. | `default` |
| `manifests` | string | The path to the manifests. Can be a file or directory. | `/cnab/app/kubernetes` |
| `force` | boolean | If true, immediately remove resources from API and bypass graceful deletion. Note that immediate deletion of some resources may result in inconsistency or data loss and requires confirmation. Sets grace period to `0`. | `false` |
| `gracePeriod` | integer | Period of time in seconds given to the resource to terminate gracefully. Ignored if negative. Set to 1 for immediate shutdown. | `-1` |
| `timeout` | integer | The length of time (in seconds) to wait before giving up on a delete, zero means determine a timeout from the size of the object. | 0 |
| `wait` | boolean | If true, wait for resources to be gone before returning. This waits for finalizers. | `true` |

### Outputs

This mixin will leverage the `kubectl get` command in order to populate outputs. Given the wide range of objects that can be created, the mixin will support JSON Path to specify how to retrieve values to populate outputs. Bundle authors will specify the object type, name and provide a JSONPath to obtain the data. The mixin will not attempt further processing of the data, so if a JSONPath expression is given that results in multiple items, the JSON representing that will be stuck into the output as is. Namespace will default to `default` if not specified

For example, to obtain the ClusterIP of a a given service, consider the following porter.yaml excerpt:

```yaml
install:
- kubernetes:
    description: "Install Super Cool App"
    manifests:  "/cnab/app/manifests/super-cool-app"
    outputs:
      - name: cluster_ip
        resourceType: "service"
        resourceName: "super-cool-service"
        namespace: "cool"
        jsonPath: "spec.clusterIP"
```