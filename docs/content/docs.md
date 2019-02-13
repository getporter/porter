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

$ duffle install myapp -f bundle.json
```

Here's a sample Porter manifest:

```yaml
mixins:
- helm

name: mydb
version: 0.1.0
invocationImage: deislabs/porter-mysql:latest

credentials:
- name: kubeconfig
  path: /root/.kube/config

install:
- description: "Install MySQL"
  helm:
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
- description: "Uninstall MySQL"
  helm:
    name: mydb
    purge: true
```

# Next Steps

* [Install Porter](/install/)
* [Quick Start](/quickstart/)