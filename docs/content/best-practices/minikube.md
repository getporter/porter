---
title: Connect to Minikube
description: How to connect to a Minikube cluster inside a Porter bundle.
---

The easiest way to work with Minikube is to embed the certificates used to authenticate to your cluster inside your kubeconfig. Minikube can do that for you automatically with a configuration setting:

```
minikube config set embed-certs true
```

With that setting enabled, the next time you run `minikube start`, your kubeconfig will have the certificates embedded.
