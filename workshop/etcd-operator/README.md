# Install a Helm Chart

Make a new bundle and install the Helm chart for etcd-operator

1. Create a porter bundle in a new directory with `porter create`.
1. Modify the **porter.yaml** to use the **helm** mixin and define credentials for **kubeconfig**.
1. Using the helm mixin, install the latest **stable/etcd-operator** chart with the default values.
1. Generate credentials for your bundle.
1. Install your bundle.
