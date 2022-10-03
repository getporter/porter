---
title: "Example: Airgapped Environments"
description: "Learn how deploy in a disconnected or airgapped environments with Porter"
weight: 10
---

The [ghcr.io/getporter/examples/whalegap] bundle demonstrates how to create a bundle for airgapped, or disconnected, environments. 

Source: https://getporter.org/examples/src/airgap

The whalegap bundle distributes the [whalesay app], which is deployed with Helm to a Kubernetes cluster.
This application serves an endpoint that draws the cowsay CLI output as a whale.

Let's walk through how to create a bundle that can be deployed in an airgapped environment.

```
 _____________________
< Challenge Accepted! >
 ---------------------
    \
     \
      \
                    ##        .
              ## ## ##       ==
           ## ## ## ##      ===
       /""""""""""""""""___/ ===
  ~~~ {~~ ~~~~ ~~~ ~~~~ ~~ ~ /  ===- ~~~
       \______ o          __/
        \    \        __/
          \____\______/
```

## Author the bundle
The bundle must declare any additional images required for installation in order for Porter to include them in the bundle.

Here's the [full example bundle][whalegap] for you to
follow along with.

[whalegap]: /examples/src/airgap/

### Declare Images

Add an [images] section and declare every image that your application relies upon:

```yaml
images:
  whalesayd:
      description: "Whalesay as a service"
      imageType: "docker"
      repository: "carolynvs/whalesayd"
      digest: "sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f"
```

The repository field should be the image name, without a tag or digest included, and the digest is the repository digest of the image.
You can get that information by running the following docker command:

```console
$ docker image inspect carolynvs/whalesayd:v0.1.0 --format '{{.RepoDigests}}'
[carolynvs/whalesayd@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f]
```

### Specify new image location

When a bundle is published, any images listed in this section are published with the bundle as well.
The new location of the image is made available as a template variable so that you can reference the new location in the bundle actions.
You must work with the image using its digest and not a tag.

```yaml
install:
  - helm3:
      description: "Install WhaleGap"
      name: whalegap
      chart: ./charts/whalegap
      replace: true
      set:
        image.repository: ${ bundle.images.whalesayd.repository }
        image.digest: ${ bundle.images.whalesayd.digest }
```

The helm chart must have been written to allow specifying the location of any images in digested form.
Then in your install step, use the helm \--set flag overriding the image with its location in the airgapped environment.
Porter handles tracking the image location for you, just use the template variables to swap the image used by the helm chart.

[images]: /author-bundles/#images

## Move the bundle across the airgap

Let's simulate moving the bundle across an airgap by publishing the bundle to a different registry.
If you _really_ want to practice an airgapped deployment, set up a [local KinD cluster with its own registry](https://kind.sigs.k8s.io/docs/user/local-registry/), and then before installing the bundle, turn off your wifi (or unplug your ethernet cable).

1. Use the archive command to create a tgz file of the bundle, which includes all images referenced in the porter.yaml file.
    ```console
     porter archive whalegap.tgz --reference ghcr.io/getporter/examples/whalegap:v0.2.0
    ```
2. If you have a disconnected network to test with, copy whalegap.tgz over to it now.
   For this example, we will do the best we can without a real airgap, and check what was deployed to verify it didn't access the original location of the images.
3. Publish the bundle from the archive file to a different registry, such as your personal Docker Hub account.
    ```console
    porter publish --archive whalegap.tgz --reference YOURNAME/whalegap:v0.1.1
    ```

## Run the bundle

Now that the bundle is ready to use in our "airgapped" environment, let's install the bundle and see what happens.

First, create a credential set that includes the target cluster's kubeconfig.
Edit the filepath to the kubeconfig with a path to a valid kubeconfig file.

```yaml
# mycluster-credentials.yaml
schemaType: CredentialSet
schemaVersion: 1.0.1
name: mycluster
credentials:
  - name: kubeconfig
    source:
      path: $HOME/.kube/config
```

Apply the credential set so that it can be used when the bundle is installed:

```console
porter credentials apply mycluster-credentials.yaml
```

Finally, install the bundle referencing the new location of the bundle:

```console
porter install airgap-example --reference YOURNAME/whalegap:v0.1.1 -c mycluster
```

After the bundle finishes installing, inspect the pod created by the bundle to verify that the pod is using the new location that is accessible from within the airgapped environment. 

```console
$ kubectl get pods
NAME                            READY   STATUS      RESTARTS   AGE
whalegap-5bf4dcdb86-zb4rm       1/1     Running     0          40s

$ kubectl describe pod whalegap-5bf4dcdb86-zb4rm
Name:         whalegap-5bf4dcdb86-zb4rm
Namespace:    quickstart
Priority:     0
Node:         porter-control-plane/172.19.0.3
Start Time:   Mon, 07 Feb 2022 12:13:46 -0600
Labels:       app.kubernetes.io/instance=whalegap
              app.kubernetes.io/name=whalegap
              pod-template-hash=5bf4dcdb86
Annotations:  <none>
Status:       Running
IP:           10.244.0.19
IPs:
  IP:           10.244.0.19
Controlled By:  ReplicaSet/whalegap-5bf4dcdb86
Containers:
  whalegap:
    Container ID:   containerd://029d836e5b1818b6a99b7b2783087807e4adaa12521011d2870fbaf2ed876a5e
    Image:          YOURNAME/whalegap@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f
    Image ID:       docker.io/YOURNAME/whalegap@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f
```

From the output of kubectl describe, you can see that the image deployed was not the original image referenced in the bundle, carolynvs/whalesayd, and instead it references a new location _inside_ the relocated bundle repository.
All referenced images are published into the same repository as the bundle, and they are only available using their digest, not the original image name.

## Next Steps

* [Understand how Porter publishes archived bundles to a registry](/archive-bundles/)

[ghcr.io/getporter/examples/whalegap]: https://github.com/orgs/getporter/packages/container/package/examples%2Fwhalegap
[whalesay app]: https://github.com/carolynvs/whalesayd
