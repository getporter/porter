---
title: "Porter Operator v1.0.0 is here!"
description: "Announcing the Porter Operator's v1.0.0 release"
date: "2023-09-28"
authorname: "Sarah Christoff"
author: "@schristoff"
tags: ["operator", "v1"]
---

Porter Operator v1.0.0 is finally here! 🎉

## What is the Porter Operator?

The Porter Operator gives you a native, integrated experience for managing your bundles from Kubernetes. It is the recommended way to automate your bundle pipeline with support for GitOps.

## What's new in v1?
Our last release was v0.8.2, which updated a lot of dependencies.

### Outputs

Our biggest feature by far is the Outputs API. Being able to share outputs between bundles,
services, and allow users to retrieve outputs from bundles is a must have. By creating the
`installation.outputs` CRD we are now able to expose those outputs. We leverage the work done
in Porter to turn a gRPC server which enables for these outputs to be shared for a variety of
use cases. 

### Delete all the things!

There were a myriad of issues we were running into when trying to delete resources. One of these issues being that a finalizer would block the deletion of a namespace. By updating one of our reconcilers to check that a deletion timestamp was set, the deletion of namespaces which have operator resources is now able to occur.

The `TTLSecondsAfterFinished` agent configuration option is out, which allows users to set the time limit of a job that has finished its execution. This defaults to 600 seconds, but in the long term should allow for easier debugging.

### Checkout the Operator Demo
If you'd like to take the Operator for a spin please go through out [Quickstart here](https://porter.sh/docs/operator/quickstart/). However, if you're looking for something with more fun tools included, [Brian DeGeeter](https://github.com/bdegeeter) has this amazing [porter-argo-demo](https://github.com/bdegeeter/porter-argo-demo).


## Recognizing our key contributors
None of this would have been possible without the following:

[Carolyn Van Slyck](https://github.com/carolynvs)

[Yingrong Zhao](https://github.com/VinozzZ)

[Brian DeGetter](https://github.com/bdegeeter)

[Steven Gettys](https://github.com/sgettys)

[Troy Connor](https://github.com/troy0820)



