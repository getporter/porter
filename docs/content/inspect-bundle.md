---
title: Inspecting Bundles
description: Inspect a bundle to see the images, params, credentials, outputs and actions
---

You've found a bundle that you'd like to use, but you'd like to know information about the bundle, including what images will be used after you install the bundle. You can use the `porter inspect` command to see all the relevant information about the bundle. The `porter inspect` command will show everything that the [explain](/examining-bundles) command will show, but will also show you the invocation images and any referenced images in the bundle.

```bash
$ porter inspect --tag jeremyrickard/porter-do-bundle:v0.5.0
Name: spring-music
Description: Run the Spring Music Service on Kubernetes and Digital Ocean PostgreSQL
Version: 0.5.0

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
porter-debug    Print debug information from Porter when executing the bundle   boolean   false               false      All Actions
region          Region to create Database and DO Space                          string    nyc3                false      All Actions
space_name      Name for DO Space                                               string    jrrportertest       false      All Actions

Outputs:
Name         Description                                Type     Applies To
service_ip   IP Address assigned to the Load Balancer   string   All Actions

No custom actions defined

Invocation Images:
Image                                                                                                    Type     Digest
jeremyrickard/porter-do-bundle@sha256:74b8622a8b7f09a6802a3fff166c8d1827c9e78ac4e4b9e71e0de872fa5077be   docker   sha256:74b8622a8b7f09a6802a3fff166c8d1827c9e78ac4e4b9e71e0de872fa5077be

Images:
Name           Type     Image                                                                                                    Digest
spring-music   docker   jeremyrickard/porter-do-bundle@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f   sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f
```

With the image information above, you can use existing tooling to pull, inspect and vet the images before you run the bundle. If you copy or archive and then republish a bundle, the image information will reflect the new locations of the images, allowing you to compare between the source and the new bundle as well.

`porter inspect` can be used with a published bundle, as show above, or with a local bundle. The command even works with bundles that were not built with Porter, through the use of the `--cnab-file` flag. You can view the output in tabular form, as above, JSON or YAML. For all the options, run the command `porter inspect --help`.
