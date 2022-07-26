---
title: Examine Bundles
description: Figure out how to use a bundle
aliases:
- /examining-bundles/
---

Once a bundle has been built, how do users of the bundle figure out how to actually _use_ it? A user could read the `porter.yaml` or the `bundle.json` if they have the bundle locally, but this won't work for a bundle that has been published to an OCI registry. Even when you have them locally, the `bundle.json` and `porter.yaml` aren't the best way to figure out how to use a bundle. How should a user examine the bundle then? Porter has a command called `explain` to help with this!

```bash
$ porter explain jeremyrickard/porter-do-bundle:v1.0.0
Name: spring-music
Description: Run the Spring Music Service on Kubernetes and Digital Ocean PostgreSQL
Version: 1.0.0

Credentials:
Name               Description                                       Required
do_access_token    Access Token for Digital Ocean Account            true
do_spaces_key      DO Spaces API Key                                 true
do_spaces_secret   DO Spaces API Secret                              true
kubeconfig         Kube config file with permissions to deploy app   true

Parameters:
Name            Description                                                     Type      Default             Required   Applies To
database_name   Name of database to create                                      string    jrrportertest       false      All Actions
helm_release    Helm release name                                               string    spring-music-helm   false      All Actions
namespace       Namespace to install Spring Music app                           string    default             false      All Actions
node_count      Number of database nodes                                        integer   1                   false      All Actions
region          Region to create Database and DO Space                          string    nyc3                false      All Actions
space_name      Name for DO Space                                               string    jrrportertest       false      All Actions

Outputs:
Name         Description                                Type     Applies To
service_ip   IP Address assigned to the Load Balancer   string   install,upgrade
```

The `porter explain` command will show what credentials and parameters are required for the bundle, what outputs are generated, and what custom actions have been defined. For `parameters`, this command will also show you the default value, if one has been provided. Additionally, the user can quickly see what actions a `parameter` or `output` apply to.

`porter explain` can be used with a published bundle, as show above, or with a local bundle. The command even works with bundles that were not built with Porter, through the use of the `--cnab-file` flag. For all the options, run the command `porter explain --help`.

If you would like to see the invocation images and/or the images the bundle will use, see the [inspect](/inspect-bundles) command.
