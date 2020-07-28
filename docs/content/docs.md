---
title: Porter Docs
description: All the magic of Porter explained
---

Porter takes the work out of creating CNAB bundles. It provides a declarative authoring 
experience that lets you to reuse existing bundles, and understands how to translate 
CNAB actions to Helm, Terraform, Azure, etc.

```console
$ porter create
created porter.yaml

$ porter build
created Dockerfile
created bundle.json
created cnab/app/run

$ porter install
```

Here's a sample Porter manifest:

```yaml
mixins:
- helm

name: mysql
version: 0.1.0
tag: getporter/mysql

credentials:
- name: kubeconfig
  path: /root/.kube/config

install:
- helm:
    description: "Install MySQL"
    name: mydb
    chart: stable/mysql
    version: 0.10.2
    replace: true
    set:
      mysqlDatabase: mydb
    outputs:
    - name: "MYSQL_HOST"
      key: "MYSQL_HOST"

uninstall:
- helm:
    description: "Uninstall MySQL"
    releases:
    - mydb
    purge: true
```

# Next Steps

* [Install Porter](/install/)
* [Quick Start](/quickstart/)
* [Frequently Asked Questions](/faq)