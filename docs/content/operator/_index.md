---
title: Porter Operator
description: Automate Porter on Kubernetes with the Porter Operator
---

We are currently working on creating a Kubernetes operator for Porter.
With Porter Operator, you define installations, credential sets and parameter sets in custom resources on a cluster, and the operator handles executing Porter when the desired state of an installation changes.

The operator is not ready to use.
The initial prototype gave us a lot of feedback for how to improve Porter's support for desired state, resulting in the new [porter installation apply] command.
We are currently rewriting the operator to make use of this new command and desired state patterns.

You can watch the https://github.com/getporter/operator repository to know when new releases are ready, and participate in design discussions.

[porter installation apply]: /cli/porter_installations_apply/