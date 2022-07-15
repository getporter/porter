# Play with Google Cloud VMs

This example creates an Google Cloud VM, labels it and then deletes the test VM.

Install the gcloud mixin

```
porter mixin install gcloud
```

# Credentials

You will need a service account service key for a service account that you have
created with the Compute Instance Admin Role.

```
$ porter credentials generate gcloud
```

Select `path` and enter the full path to your service key file. This is what your credentials file should look like:

```
$ cat ~/.porter/credentials/gcloud.yaml

name: gcloud
credentials:
- name: gcloud-key-file
  source:
    path: /Users/carolynvs/Downloads/porter-test-gcloud.json
```

# Try it out

## Create a VM
```console
$ porter install --credential-set gcloud
```

## Label a VM
```console
$ porter upgrade --credential-set gcloud
```

## Delete a VM
```console
$ porter uninstall --credential-set gcloud
```
