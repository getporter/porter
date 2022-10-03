---
title: Connect to KinD
description: How to connect to a KinD cluster inside a Porter bundle.
---

[KinD](https://github.com/kubernetes-sigs/kind) clusters run Kubernetes via Docker containers.  As long as you have access to a Docker daemon, you can run KinD!

Check out the [docs](https://github.com/kubernetes-sigs/kind#installation-and-usage) to install and interact with the `kind` CLI, if you haven't already.

By default, KinD sets up the Kubernetes API server IP to be the local loopback address (`127.0.0.1`).  This is fine for interacting with the cluster from the context of your host, but when it comes to talking to the cluster from inside of another Docker image, as is the case when running an action on a Porter bundle using the default Docker driver, further configuration is needed to make this communication possible.

There are two main options:

1. Set up public DNS to map to the default API server address.  This would ideally include ingress TLS and so presents a secure option for communication, but does involve a larger overhead for setup.
1. Configure the KinD cluster to use an IP address that is already resolvable from within Porter's Docker container.  This will most likely lead to considerable security implications and is not advised for clusters hosting actual workloads or sensitive information.

Here we'll look at the latter option of configuring the KinD cluster to use a [different API Server address](https://kind.sigs.k8s.io/docs/user/configuration/#api-server).

### KinD Setup

First, we need to determine our host IP address.  I'll take the example of a Mac OSX host:

```console
$ ifconfig en0 | awk '/inet / {print $2; }' | cut -d ' ' -f 2
10.0.1.4
```

We can then use this IP in the [KinD config](https://kind.sigs.k8s.io/docs/user/configuration):

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  apiServerAddress: 10.0.1.4
  apiServerPort: 6443
```

After saving the config to `kind-config.yaml`, we can create our cluster:

```console
$ kind create cluster --config kind-config.yaml
```

Once the cluster is successfully created, the generated kubeconfig should be merged into the default location (`~/.kube/config` or the location specified by the `KUBECONFIG` env var).  We can test API server communications with `kubectl`:

```console
 $ kubectl cluster-info
Kubernetes master is running at https://10.0.1.4:6443
KubeDNS is running at https://10.0.1.4:6443/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```

### Porter time!

Porter bundles that access a Kubernetes cluster when running can now be installed as normal, once the [Credential Set](../credentials) is generated/edited to use the KinD kubeconfig.

Here we'll create and edit credentials and then install the [MySQL bundle](/src/build/testdata/bundles/mysql):

```console
 $ porter credentials create mysql.json
creating porter credential set in the current directory
 $ cat mysql.json
# modify mysql.json with your editor to the content below
{
    "schemaType": "CredentialSet",
    "schemaVersion": "1.0.1",
    "name": "mysql",
    "credentials": [
        {
            "name": "kubeconfig",
            "source": {
                "path": "/Users/vdice/.kubes/kind/kubeconfig"
            }
        }
    ]
}
 $ porter cdredentials apply mysql.json
Applied /mysql credential set
 $ porter install -c mysql
installing mysql...
executing install action from mysql (bundle instance: mysql)
Install MySQL
...
/usr/local/bin/helm helm install --name porter-ci-mysql bitnami/mysql --version 6.14.2 --replace --set db.name=mydb --set db.user=mysql-admin
NAME:   porter-ci-mysql
LAST DEPLOYED: Wed Jul 15 15:11:43 2020
NAMESPACE: default
STATUS: DEPLOYED
...
execution completed successfully!
```
