---
title: Connect to Minikube
description: How to connect to a Minikube cluster inside a Porter bundle.
---

The easiest way connect to a Minikube cluster from inside a Porter bundle is to
embed the certificates used to authenticate to your cluster inside your
kubeconfig. Minikube can do that for you automatically with a configuration
setting:

```
minikube config set embed-certs true
```

With that setting enabled, the next time you run `minikube start`, your
kubeconfig will have the certificates embedded.

🚨 There is an open issue with using Minikube's Docker daemon when _building_ a
bundle [#1383](https://github.com/getporter/porter/issues/1383). This article
only describes how to use a Minikube kubeconfig inside a running bundle.
