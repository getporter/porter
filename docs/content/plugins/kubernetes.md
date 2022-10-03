---
title: Kubernetes Plugin
description: Integrate Porter with Kubernetes
---

<img src="/images/mixins/kubernetes.svg" class="mixin-logo" style="width: 300px"/>

Integrate Porter with Kubernetes.

Source: https://github.com/getporter/kubernetes-plugins

## Install or Upgrade

The plugin is distributed as a single binary, `kubernetes`.
The following snippet will clone this repository, build the binary and install it to **~/.porter/plugins/**.

```
porter plugin install kubernetes --version v1.0.0-rc.1
```

Note that the v1 release of the plugin only works with Porter v1.0.0-alpha.20 and higher.

## Usage

After installation, you must modify your porter configuration file, see an
walkthrough for the `kubernetes.secrets` plugin configuration below.

The plugin can be used when porter is running inside a Kubernetes cluster - in which case it will connect automatically, it can also be used from outside a cluster in which case it will either use the kubeconfig file sourced from the `KUBECONFIG` environment variable or `$HOME/.kube/config` if this is not set.

When running outside a cluster the plugin requires configuration to specify which namespace it should store data in, when running inside a cluster it will use the namespace of the pod that porter is running in.

The plugin also requires that the user or service account that is being used with Kubernetes has `"get","list","create","delete",` and `"patch"` permissions on secrets in the namespace.

The [Porter Operator](https://github.com/getporter/operator) is the primary use case
for running in Kubernetes which configures the necessary service accounts via 
it's `configureNamespace` custom action.

```
porter invoke porterops --action configureNamespace --param namespace=quickstart -c porterops
```

## Plugin Configuration

### Secrets

The `kubernetes.secrets` plugin enables resolution of credential or parameter values and storing sensitive data(parameter or output values) as secrets in Kubernetes via the Porter Operator.

1. Create, `./porter-k8s-config.yaml`
1. Add the following lines1:

    ```yaml
    default-secrets: "kubernetes-secrets"
    secrets:
    - name: "kubernetes-secrets"
      plugin: "kubernetes.secrets"
    ```
1. Provide the Porter config to the `configureNamespace` operator bundle action

    ```
    porter invoke operator --action=configureNamespace --param namespace=<namespace name> --param porterConfig=porter-k8s-config.toml -c kind -n=operator
    ```

* If the plugin is being used outside of a Kubernetes cluster then add the following lines to specify the namespace to be used to store data:

    ```yaml
    default-secrets: "kubernetes-secrets"
    secrets:
      - name: "kubernetes-secrets"
        plugin: "kubernetes.secrets"
        config:
          namespace: "<namespace name>"
    ```

In both cases the Kubernetes secret must be created with a `credential` key
```
kubectl --namespace "<namespace name>" create secret generic password --from-literal=credential=test 
```

Porter credentials file `test-credentials.yaml`
```
---
schemaType: CredentialSet
schemaVersion: 1.0.1
namespace: ''
name: kubernetes-plugin-test
credentials:
- name: test-cred
  source:
    secret: password
```

```
porter credentials apply test-credentials.yaml
```
