---
title: Kubernetes plugin
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

### Storage

`Kubernetes.storage` plugin enables Porter to store data, such as claims, parameters and credentials, in a Kubernetes cluster.The plugin stores data in Kubernetes as secrets.

1. Open, or create, `~/.porter/config.toml`.
2. Add the following lines:
   
   ```toml
   default-storage = "kubernetes-storage"

   [[storage]]
   name = "kubernetes-storage"
   plugin = "kubernetes.storage" 
   ```
3. If the plugin is being used outside of a Kubernetes cluster then add the following lines to specify the namespace to be used to store data:

   ```toml
   [storage.config]
   namespace = "<namespace name>"
   ```

### Secrets

`Kubernetes.secrets` plugin enables resolution of credential or parameter values as secrets in Kubernetes.

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

### Storage and Secrets combined

When both storage and secrets are configured, be sure to place the `default-*` stanzas at the top of the file, like so:

```toml
default-secrets = "kubernetes-secrets"
default-storage = "kubernetes-storage"

[[secrets]]
name = "kubernetes-secrets"
plugin = "kubernetes.secret"

[[storage]]
name = "kubernetes-storage"
plugin = "kubernetes.storage"
```

If runing outside of Kubernetes then also include the namespace configuration

```toml
[secrets.config]
namespace = "<namespace name>"

[storage.config]
namespace = "<namespace name>"
```