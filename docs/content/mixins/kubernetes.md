---
title: kubernetes mixin
description: Manage a set of Kubernetes manifests using the kubectl CLI
---

<img src="/images/mixins/kubernetes.svg" class="mixin-logo" style="width: 150px"/>

Manage a set of Kubernetes manifests using the [kubectl CLI](https://kubernetes.io/docs/reference/kubectl/).

Source: https://github.com/getporter/kubernetes-mixin

### Install or Upgrade

```shell
porter mixin install kubernetes --version v1.0.0-rc.2
```

### Install or Upgrade canary version

```shell
porter mixin install kubernetes --version canary --url https://cdn.porter.sh/mixins/kubernetes
```

### Install or Upgrade from feed-url

```shell
porter mixin install kubernetes --feed-url https://cdn.porter.sh/mixins/atom.xml
```

### Manually Install or Upgrade with a specific version from github

```shell
porter mixin install kubernetes --version $VERSION --url https://github.com/getporter/kubernetes-mixin/releases/download
```

### Examples

### Mixin Configuration

#### Kubernetes client version

```yaml
- kubernetes:
    clientVersion: v1.15.5
```

### Mixin Actions Syntax

#### Install Action

```yaml
install:
  - kubernetes:
      description: "Install Hello World App"
      manifests:
        - /cnab/app/manifests/hello
      wait: true

```

#### Install Upgrade Action

```yaml
upgrade:
  - kubernetes:
      description: "Upgrade Hello World App"
      manifests:
        - /cnab/app/manifests/hello
      wait: true

```

#### Uninstall Action

```yaml
uninstall:
  - kubernetes:
      description: "Uninstall Hello World App"
      manifests:
        - /cnab/app/manifests/hello
      wait: true

```

#### Outputs

The mixin supports extracting resource metadata from Kubernetes as outputs.

```yaml
outputs:
    - name: NAME
      resourceType: RESOURCE_TYPE
      resourceName: RESOURCE_TYPE_NAME
      namespace: NAMESPACE
      jsonPath: JSON_PATH_DEFINITION
```
