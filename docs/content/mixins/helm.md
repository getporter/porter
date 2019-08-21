---
title: helm mixin
description: Using the helm mixin
---

<img src="/images/mixins/helm.svg" class="mixin-logo" style="width: 150px"/>

Manage a Helm release

Source: https://github.com/deislabs/porter-helm

### Install or Upgrade
```
porter mixin install helm --feed-url https://cdn.deislabs.io/porter/atom.xml
```

### Examples

Install

```yaml
install:
- helm:
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
      secret: "{{ bundle.parameters.mysql-name }}"
      key: mysql-root-password
    - name: mysql-password
      secret: "{{ bundle.parameters.mysql-name }}"
      key: mysql-password
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