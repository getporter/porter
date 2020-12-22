---
title: helm3 mixin
description: Manage a Helm release with the helm v3 CLI
---

<img src="https://porter.sh/images/mixins/helm.svg" align="right" width="150px"/>

This is a Helm3 mixin for [Porter](https://github.com/getporter/porter). It executes the
appropriate helm command based on which action it is included within: `install`,
`upgrade`, or `delete`.

Source: https://github.com/MChorfa/porter-helm3

### Install or Upgrade

```shell
porter mixin install helm3 --feed-url https://mchorfa.github.com/porter-helm3/atom.xml
```

### Mixin Configuration

- Helm client version configuration. You can define others minors and patch versions up and down

```yaml
- helm3:
    clientVersion: v3.3.4
```

Repositories

```yaml
- helm3:
    repositories:
      stable:
        url: "https://charts.helm.sh/stable"
```

### Mixin Syntax

Install

```yaml
install:
  - helm3:
      description: "Description of the command"
      name: RELEASE_NAME
      chart: STABLE_CHART_NAME
      version: CHART_VERSION
      namespace: NAMESPACE
      replace: BOOL # Remove it if upsert is set to true. This is unsafe in production
      devel: BOOL
      wait: BOOL # default true
      upsert: BOOL # default false. If set to true `upgrade --install` will be executed
      set:
        VAR1: VALUE1
        VAR2: VALUE2
```

Upgrade

```yaml
install:
  - helm3:
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
  - helm3:
      description: "Description of command"
      namespace: NAMESPACE
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
  - helm3:
      description: "Install MySQL"
      name: mydb
      chart: stable/mysql
      version: 0.10.2
      namespace: mydb
      replace: true
      set:
        mysqlDatabase: wordpress
        mysqlUser: wordpress
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
  - helm3:
      description: "Upgrade MySQL"
      name: porter-ci-mysql
      chart: stable/mysql
      version: 0.10.2
      wait: true
      resetValues: true
      reuseValues: false
      set:
        mysqlDatabase: mydb
        mysqlUser: myuser
        livenessProbe.initialDelaySeconds: 30
        persistence.enabled: true
```

Uninstall

```yaml
uninstall:
  - helm3:
      description: "Uninstall MySQL"
      namespace: mydb
      releases:
        - mydb
```
