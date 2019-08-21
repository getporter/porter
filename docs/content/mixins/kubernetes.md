---
title: kubernetes mixin
description: Using the kubernetes mixin
---

<img src="/images/mixins/kubernetes.svg" class="mixin-logo" style="width: 150px"/>

Apply a set of Kubernetes manifests

Source: https://github.com/deislabs/porter/tree/master/pkg/kubernetes

### Install or Upgrade
```
porter mixin install kubernetes --feed-url https://cdn.deislabs.io/porter/atom.xml
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