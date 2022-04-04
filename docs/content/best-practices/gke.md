---
title: Connect to GKE
description: How to connect to a GKE cluster inside a Porter bundle.
---

GKE cluster authentication requires more than just a kubeconfig, it also needs a
service account configured.

See the [GKE example][example] for a full working example bundle.

[example]: https://porter.sh/src/examples/gke-example

1. [Generate a kubeconfig](#generate-a-kubeconfig).
1. [Create a service account](#create-a-service-account).
1. Define credentials in **porter.yaml** for the kubeconfig 
    and service account:

    ```yaml
    credentials:
        - name: kubeconfig
          path: /home/nonroot/.kube/config
        - name: google-service-account
          path: /home/nonroot/google-service-account.json
    ```

1. Define an environment variable, `GOOGLE_APPLICATION_CREDENTIALS` that
   contains the path to the service account file,
   `/home/nonroot/google-service-account.json`.

    This can be accomplished via one of the methods below. The first method is
    recommended over using a parameter. Using parameters to define environment
    variables is a hack provided only for the purpose of this example.

    * Add the following line to your [Custom Dockerfile](/bundle/custom-dockerfile):

        ```
        ENV GOOGLE_APPLICATION_CREDENTIALS=/home/nonroot/google-service-account.json
        ```
    * Add a parameter to **porter.yaml**:

        ```yaml
        parameters:
            - name: google-app-creds
              env: GOOGLE_APPLICATION_CREDENTIALS
              default: /home/nonroot/google-service-account.json
        ```

---

## Generate a kubeconfig
1. You must have `gcloud` installed locally, and be authenticated.
1. Define the following environment variables:

    ```bash
    CLUSTER="REPLACE_WITH_YOUR_CLUSTER_NAME"
    ZONE="REPLACE_WITH_YOUR_CLUSTER_ZONE"
    PROJECT="REPLACE_WITH_YOUR_GOOGLE_PROJECT"
    GET_CMD="gcloud container clusters describe $CLUSTER --zone=$ZONE --project=$PROJECT"
    ```
1. Run the following command to create a kubeconfig for your GKE cluster:

        cat > kubeconfig.yaml <<EOF
        apiVersion: v1
        kind: Config
        current-context: my-cluster
        contexts: [{name: my-cluster, context: {cluster: cluster-1, user: user-1}}]
        users: [{name: user-1, user: {auth-provider: {name: gcp}}}]
        clusters:
        - name: cluster-1
        cluster:
            server: "https://$(eval "$GET_CMD --format='value(endpoint)'")"
            certificate-authority-data: "$(eval "$GET_CMD --format='value(masterAuth.clusterCaCertificate)'")"
        EOF

1. Move the kubeconfig.yaml to a location where you would like to keep it,
   for example `$HOME/.kube/my-gke-cluster.yaml`.

This file contains your master's IP address and the cluster's CA certificate but
does not contain enough information to authenticate to the cluster.

---

## Create a service account

1. [Create a service account][sa].
1. [Assign the account access to GKE][iam], such as Kubernetes Engine
   Developer.
1. Create a service account key file, e.g. service-account.json, and save the
   file locally.

This is a sensitive file that contains enough information to perform actions
against your Google account. Keep it safe. 🔐

[sa]: https://cloud.google.com/iam/docs/creating-managing-service-accounts
[iam]: https://cloud.google.com/kubernetes-engine/docs/how-to/iam

A big thanks to https://ahmet.im/blog/authenticating-to-gke-without-gcloud/ for
helping us figure out how to authenticate to GKE properly! 🙇‍♀️

