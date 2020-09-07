---
title: kubernetes mixin
description: Manage a set of Kubernetes manifests using the kubectl CLI
---

<img src="/images/mixins/kubernetes.svg" class="mixin-logo" style="width: 150px"/>

Manage a set of Kubernetes manifests using the [kubectl CLI](https://kubernetes.io/docs/reference/kubectl/).

Source: https://github.com/deislabs/porter-kubernetes

### Install or Upgrade

```shell
porter mixin install kubernetes
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
porter mixin install kubernetes --version $VERSION --url https://github.com/deislabs/porter-kubernetes/releases/download
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
