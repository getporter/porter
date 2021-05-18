---
title: "View Logs From Previous Runs"
description: |
    Porter now saves the logs from a run and you can view the logs from previous runs.
date: "2021-03-22"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://twitter.com/carolynvs"
authorimage: "https://github.com/carolynvs.png"
tags: ["release-notes"]
---

With the [v0.35.0 release of Porter][v0.35.0], the logs generated when
install/upgrade/invoke/uninstall is run are persisted. Now you can view logs
from previous runs, which can come in handy when troubleshooting a deployment.

You can view the logs from the most recent execution of a bundle with the
`porter logs` command:

```bash
$ porter logs -i whalegap
executing install action from whalegap (installation: whalegap)
Install WhaleGap
/usr/local/bin/helm helm install --name whalegap ./charts/whalegap --replace --set image.digest=sha256:5cca9dfa8ba540a32537d586651d3918d6f39761cdf4457fbe32c58c36c1defc --set image.repository=carolynvs/trust-demo --set msg=whale hello there!
NAME:   whalegap
LAST DEPLOYED: Tue Mar  9 16:43:15 2021
NAMESPACE: porter-operator-system
STATUS: DEPLOYED

RESOURCES:
==> v1/Deployment
NAME      READY  UP-TO-DATE  AVAILABLE  AGE
whalegap  0/1    1           0          0s

==> v1/Pod(related)
NAME                       READY  STATUS             RESTARTS  AGE
whalegap-76b77b9f99-6ww22  0/1    ContainerCreating  0         0s

==> v1/Service
NAME      TYPE       CLUSTER-IP     EXTERNAL-IP  PORT(S)  AGE
whalegap  ClusterIP  10.108.26.188  <none>       80/TCP   0s

==> v1/ServiceAccount
NAME      SECRETS  AGE
whalegap  1        0s


NOTES:
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace porter-operator-system -l "app.kubernetes.io/name=whalegap,app.kubernetes.io/instance=whalegap" -o jsonpath="{.items[0].metadata.name}")
  echo "Visit http://127.0.0.1:8080 to use your application"
  kubectl port-forward $POD_NAME 8080:80

execution completed successfully!
```

We have changed the output of the `porter show` command to provide IDs for
previous executions of a bundle:

```bash
$ porter show whalegap
Name: whalegap
Created: 2021-01-27
Modified: 2021-03-09

Outputs:
-------------------------------------------------------------------------------
  Name                                 Type    Value
-------------------------------------------------------------------------------
  io.cnab.outputs.invocationImageLogs  string  executing install action from
                                               whalegap (installation: wha...

History:
--------------------------------------------------------------------------
  Run ID                      Action     Timestamp   Status     Has Logs
--------------------------------------------------------------------------
  01EX0JS7JX8K2A4JQ8NAABVBDZ  install    2021-01-27  succeeded  false
  01F0BXXXE9MSJZRMW0M95V4DF9  install    2021-03-09  succeeded  true
```

In the output above, we can see that I first tried to install a bundle and it
failed. Since it was executed with a verison of Porter older than v0.35.0, no
logs are available. Then I reinstalled the bundle using the latest version of
Porter, and luckily this time it succeeded, and the logs were persisted.

You can view the logs from a specific run using the `--run` flag, which is
really useful when you are figuring out how to capture text output when the
bundle was run, or comparing a good/bad run to figure out what went wrong.

```bash
porter logs --run 01F0BXXXE9MSJZRMW0M95V4DF9
```

Happy logging! 🔎

[v0.35.0]: https://github.com/getporter/porter/releases/v0.35.0
