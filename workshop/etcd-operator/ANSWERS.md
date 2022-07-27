# Install a Helm Chart

Make a new bundle and install the Helm chart for etcd-operator

1. Create a porter bundle in a new directory with `porter create`.
    ```console
    $ mkdir etcd-operator
    $ cd etcd-operator
    $ porter create
    ```
1. Modify the **porter.yaml** to use the **helm** mixin and define credentials for **kubeconfig**. 
1. Using the helm mixin, install the latest **stable/etcd-operator** chart with the default values. See [porter.yaml](porter.yaml) for the finished manifest.
1. Generate credentials for your bundle.
    ```console
    $ porter credentials generate
    ```
    
    See [etcd-operator.yaml](etcd-operator.yaml) for what it should look like in **~/.porter/credentials/etcd-operator.yaml**.
1. Install your bundle.
    ```console
    $ porter install --credential-set etcd-operator
    ```