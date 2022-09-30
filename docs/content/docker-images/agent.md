---
title: Porter Agent Docker Image
description: The Porter Operator agent image
---

The [ghcr.io/getporter/porter-agent][porter-agent] Docker image is intended for use by the [Porter Operator] which runs on Kubernetes.
If you need to run Porter in a local container, not on Kubernetes, you should use the [porter client] image.
The porter agent is also available in the [PlatformOne IronBank registry](https://registry1.dso.mil/harbor/projects/3/repositories/opensource%2Fgetporter%2Fporter-agent/artifacts-tab).

It has tags that match what is available from our [install](/install/) page: latest, canary and specific versions such as v0.38.1.

The [configuration] file for Porter should be mounted in a volume to **/porter-config**.
The image will copy the configuration file into PORTER_HOME when the container starts and then run the specified porter command, similar to the [porter client] image.

## Example

This set of manifests performs the follow actions:
1. Create a namespace named porter-agent-test.
1. Create a role named porter-agent-role with sufficient permissions to run Porter.
1. Create a service account named porter-agent and add it to the porter-agent-role.
1. Create a persistent volume claim named porter-hello-shared that Porter uses to share data with the bundle's pod.
1. Create a pod named porter-hello-3591 that executes the install action for the ghcr.io/getporter/examples/porter-hello:v0.2.0 bundle using the `kubernetes` driver.
   The kubernetes driver executes the bundle in a pod on a Kubernetes cluster.

Run the following command to run the porter-hello bundle on a cluster to try it out.

```
kubectl apply -f https://raw.githubusercontent.com/getporter/porter/a059a9668934dff475f9d9633781d2f32512581d/examples/porter-agent-manifest.yaml
```

<script src="https://gist-it.appspot.com/https://github.com/getporter/porter/blob/main/examples/porter-agent-manifest.yaml"></script>

[configuration]: /configuration
[porter-agent]: https://github.com/getporter/porter/pkgs/container/porter-agent
[porter client]: /docker-images/client/
[Porter Operator]: https://github.com/getporter/operator
