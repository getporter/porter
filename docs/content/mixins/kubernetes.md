---
title: kubernetes mixin
description: Manage a set of Kubernetes manifests using the kubectl CLI
---

<img src="/images/mixins/kubernetes.svg" class="mixin-logo" style="width: 150px"/>

Manage a set of Kubernetes manifests using the [kubectl CLI](https://kubernetes.io/docs/reference/kubectl/).

Source: https://github.com/deislabs/porter/tree/main/pkg/kubernetes

### Install or Upgrade
```
porter mixin install kubernetes
```

### Examples

```yaml
install:
  - kubernetes:
      description: "Install Hello World App"
      manifests:
        - ./manifests/hello
      wait: true
```

```yaml
uninstall:
  - kubernetes:
      description: "Uninstall Hello World App"
      manifests:
        - ./manifests/hello
      wait: true
```
