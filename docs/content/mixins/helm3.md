---
title: helm3 mixin
description: Manage a Helm release with the helm v3 CLI
aliases:
- /mixins/helm/
---

<img src="/images/mixins/helm.svg" class="mixin-logo" style="width: 150px"/>

This is a [Helm](https://helm.sh) v3 mixin for
[Porter](https://github.com/getporter/porter). It executes the appropriate helm v3
command based on which action it is included within: `install`, `upgrade`, or
`delete`.

Source: https://github.com/MChorfa/porter-helm3

### Install or Upgrade

Currently, we only support the installation via `--feed-url`. Please make sure to install the mixin as follow:

```shell
porter mixin install helm3 --version v1.0.0-rc.2 --feed-url https://mchorfa.github.io/porter-helm3/atom.xml
```

### Mixin Configuration

Helm client version configuration. You can define others minors and patch versions up and down

```yaml
- helm3:
    clientVersion: v3.7.0
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
      devel: BOOL
      wait: BOOL # default true
      noHooks: BOOL # disable pre/post upgrade hooks (default false)
      skipCrds: BOOL # if set, no CRDs will be installed (default false)
      timeout:  DURATION # time to wait for any individual Kubernetes operation
      debug: BOOL # enable verbose output (default false)
      set:
        VAR1: VALUE1
        VAR2: VALUE2
      values: # Array of paths to: Set/Override multiple values and multi-lines values
        - PATH_TO_THE_VALUES_FILE_1
        - PATH_TO_THE_VALUES_FILE_2
        - PATH_TO_THE_VALUES_FILE_3
```

Upgrade

```yaml
upgrade:
  - helm3:
      description: "Description of the command"
      name: RELEASE_NAME
      chart: STABLE_CHART_NAME
      version: CHART_VERSION
      namespace: NAMESPACE
      resetValues: BOOL
      reuseValues: BOOL
      wait: BOOL # default true
      noHooks: BOOL # disable pre/post upgrade hooks (default false)
      skipCrds: BOOL # if set, no CRDs will be installed (default false)
      timeout:  DURATION # time to wait for any individual Kubernetes operation
      debug: BOOL # enable verbose output (default false)
      set:
        VAR1: VALUE1
        VAR2: VALUE2
      values: # Array of paths to: Set/Override multiple values and multi-line values
        - PATH_TO_THE_VALUES_FILE_1
        - PATH_TO_THE_VALUES_FILE_2
        - PATH_TO_THE_VALUES_FILE_3
```

Uninstall

```yaml
uninstall:
  - helm3:
      description: "Description of command"
      namespace: NAMESPACE
      releases:
        - RELEASE_NAME1
        - RELEASE_NAME2
      wait: BOOL # default false, if set It will wait for as long as --timeout
      noHooks: BOOL # prevent hooks from running during uninstallation
      timeout:  DURATION # time to wait for any individual Kubernetes operation
      debug: BOOL # enable verbose output (default false)
```

#### Outputs

The mixin supports saving secrets from Kubernetes as outputs.

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
      skipCrds: true
      set:
        mysqlDatabase: wordpress
        mysqlUser: wordpress
      values:
        - "./manifests/values_1.yaml"
        - "./manifests/values_2.yaml"
        - "./manifests/values_3.yaml"
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
      noHooks: true
      set:
        mysqlDatabase: mydb
        mysqlUser: myuser
        livenessProbe.initialDelaySeconds: 30
        persistence.enabled: true
      values:
        - "./manifests/values_1.yaml"
        - "./manifests/values_2.yaml"
        - "./manifests/values_3.yaml"
```

Uninstall

```yaml
uninstall:
  - helm3:
      description: "Uninstall MySQL"
      namespace: mydb
      releases:
        - mydb
      wait: true
      noHooks: true
```

Execute

```yaml
login:
  - helm3:
      description: "Login to OCI registry"
      arguments:
        - registry
        - login
        - localhost:5000
        - "--insecure"
      flags:
        u: myuser
        p: mypass
```
