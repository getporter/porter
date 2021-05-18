---
title: View Logs
description: View logs from previous runs of Porter
---

You can view the logs from the most recent execution of a bundle with the
`porter logs` command. Logs are only persisted by v0.35.0+ of Porter.

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

You can use the `porter show` command to see the ID of a previous run. The **Run
ID** uniquely identifies a run, or execution of a bundle.

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

You can view the logs from a specific run using the `--run` flag, which is
really useful when you are figuring out how to capture text output when the
bundle was run, or comparing a good/bad run to figure out what went wrong.

```bash
porter logs --run 01F0BXXXE9MSJZRMW0M95V4DF9
```
