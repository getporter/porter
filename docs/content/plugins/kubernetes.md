---
title: Kubernetes Plugin
description: Integrate Porter with Kubernetes
---

<img src="/images/mixins/kubernetes.svg" class="mixin-logo" style="width: 300px"/>

Integrate Porter with Kubernetes.

Source: https://github.com/getporter/kubernetes-plugins

## Install or Upgrade

```
porter plugin install kubernetes
```

## Plugin Configuration

### Secrets

`Kubernetes.secrets` plugin enables porter to store and resolve sensitive values as secrets in Kubernetes.

1. Open, or create, `~/.porter/config.toml`
2. Add the following lines:
   
   ```toml
   default-secrets = "kubernetes-secrets"

   [[secrets]]
   name = "kubernetes-secrets"
   plugin = "kubernetes.secret"
   ```
3. If the plugin is being used outside of a Kubernetes cluster then add the following lines to specify the namespace to be used to store data:
   
   ```toml
   [secrets.config]
   namespace = "<namespace name>"
   ```
