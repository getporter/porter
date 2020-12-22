---
title: helm2 mixin
description: Manage a Helm release with the helm v2 CLI
aliases:
 - /mixins/helm/
---

<img src="/images/mixins/helm.svg" class="mixin-logo" style="width: 150px"/>

This is a Helm v2 mixin for [Porter](https://github.com/getporter/porter). It
executes the appropriate helm command based on which action it is included
within: `install`, `upgrade`, or `delete`.

🚨 [Helm v2 is deprecated](https://helm.sh/blog/helm-2-becomes-unsupported/) so
you should move to Helm v3 as soon as possible. After you [migrate to Helm
3](https://helm.sh/docs/topics/v2_v3_migration/), use the [Helm 3
mixin](/mixins/helm3/). 🚀

Source: https://github.com/getporter/helm2-mixin

### Install or Upgrade

```shell
porter mixin install helm2
```

### Mixin Configuration

Helm client version

```yaml
- helm:
    clientVersion: v2.17.0
```

Add repositories

```yaml
- helm:
    repositories:
      stable:
        url: "https://charts.helm.sh/stable
```

### Mixin Syntax

Install

```yaml
install:
- helm:
    description: "Description of the command"
    name: RELEASE_NAME
    chart: STABLE_CHART_NAME
    version: CHART_VERSION
    namespace: NAMESPACE
    replace: BOOL
    devel: BOOL
    wait: BOOL # default true
    set:
      VAR1: VALUE1
      VAR2: VALUE2
```

Upgrade

```yaml
upgrade:
- helm:
    description: "Description of the command"
    name: RELEASE_NAME
    chart: STABLE_CHART_NAME
    version: CHART_VERSION
    namespace: NAMESPACE
    resetValues: BOOL
    reuseValues: BOOL
    wait: BOOL # default true
    set:
      VAR1: VALUE1
      VAR2: VALUE2
```

Uninstall

```yaml
uninstall:
- helm:
    description: "Description of command"
    purge: BOOL
    releases:
      - RELEASE_NAME1
      - RELASE_NAME2
```

#### Outputs

The mixin supports saving secrets from Kuberentes as outputs.

```yaml
outputs:
    - name: NAME
      secret: SECRET_NAME
      key: SECRET_KEY
```

The mixin also supports extracting resource metadata from Kubernetes as outputs.

```yaml
outputs:
    - name: NAME
      resourceType: RESOURCE_TYPE
      resourceName: RESOURCE_TYPE_NAME
      namespace: NAMESPACE
      jsonPath: JSON_PATH_DEFINITION
```

### Examples

Install

```yaml
install:
- helm:
    description: "Install MySQL"
    name: mydb
    chart: bitnami/mysql
    version: 6.14.2
    namespace: mydb
    replace: true
    set:
      db.name: wordpress
      db.user: wordpress
    outputs:
      - name: mysql-root-password
        secret: mydb-mysql
        key: mysql-root-password
      - name: mysql-password
        secret: mydb-mysql
        key: mysql-password
      - name: mysql-cluster-ip
        resourceType: service
        resourceName: porter-ci-mysql-service
        namespace: "default"
        jsonPath: "{.spec.clusterIP}"
```

Upgrade

```yaml
upgrade:
- helm:
    description: "Upgrade MySQL"
    name: porter-ci-mysql
    chart: bitnami/mysql
    version: 6.14.2
    wait: true
    resetValues: true
    reuseValues: false
    set:
      db.name: mydb
      db.user: myuser
      livenessProbe.initialDelaySeconds: 30
      persistence.enabled: true
```

Uninstall

```yaml
uninstall:
- helm:
    description: "Uninstall MySQL"
    purge: true
    releases:
      - mydb
```
