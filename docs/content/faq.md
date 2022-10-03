---
title: FAQ
description: Frequently Asked Questions
---

* [What is CNAB?](#what-is-cnab)
* [Does Porter fully implement the CNAB specification?](#does-porter-fully-implement-the-cnab-specification)
* [Can I use Porter bundles with other CNAB tools?](#can-i-use-porter-bundles-with-other-cnab-tools)
* [Does Porter solve something that Ansible, Terraform, etc does not?](#does-porter-solve-something-that-ansible-terraform-etc-does-not)
* [If an upgrade fails, can I roll back?](#if-an-upgrade-fails-can-i-roll-back)
* [How do I run commands that aren't in the default invocation image?](#how-do-i-run-commands-that-aren-t-in-the-default-invocation-image)
* [Are mixins just wrappers around OS or executable calls?](#are-mixins-just-wrappers-around-os-or-executable-calls)

## What is CNAB?

CNAB stands for "Cloud Native Application Bundle". When we say "bundle", that is what
we are referring to. There is a CNAB Specification and you can learn more about
it at [cnab.io](https://cnab.io).

We like to think of bundles as "cloud installers". They handle installing not just your
application, but also the underlying infrastructure that your application depends upon.

For example, let's say that you had a Wordpress bundle, which deploys on to a
Kubernetes cluster and relies on a MySQL as a Service database. The bundle could
have everything you need inside of it:

* The Wordpress chart
* The installation bash script
* Client binaries for helm and azure

The installer would take care of:

1. Requesting credentials for an existing Kubernetes cluster.
1. Provisioning an Azure MySQL database.
1. Collecting the database credentials.
1. Installing the Wordpress chart and passing in the database credentials.

The person managing the application would only need to know `porter install` and have
credentials for a Kubernetes cluster.

## Does Porter fully implement the CNAB specification?

Porter is [committed to supporting the CNAB specification](/cnab/).
We support every released sub-specification of CNAB though there are still some in draft status,
such as the security spec, and are not supported yet.

## Can I use Porter bundles with other CNAB tools?

At the time of writing, there aren't any other publicly available, maintained CNAB tools.
Porter builds CNAB-compliant bundles, except for bundles with dependencies which is a CNAB spec that has not yet been finalized.
If someone wrote a CNAB tool that matched the spec, that tool would be able to run bundles created by Porter and similarly Porter should be able to run CNAB-compliant bundles created by another tool.

If you create a bundle that uses custom extensions to the CNAB spec, and try to run it from a tool that does not support those extensions, then the tool will let you know.

Porter is at the leading-edge of the CNAB specification, vetting improvements to the specification first in Porter before agreeing that the design is solid enough to be incorporated into one of the CNAB specifications or a custom extension.

On the flip side of this, Porter is usually one of the first tools to support changes to the CNAB Specification.
So Porter can usually run any bundle created by other CNAB tools and will always let you know if a bundle isn't fully supported by Porter.

## Does Porter solve something that Ansible, Terraform, etc does not?

Porter is an implementation of the CNAB specification. Cloud Native Application
Bundles is a different way of answering "_How do I reliably, securely deploy an
application and its infrastructure?_". It isn't replacing Ansible or Terraform
but adding some concepts on top. For example, packaging together the Terraform
binary and your Terraform scripts into an installer that can be distributed via OCI
(docker) registries or USB sticks to get into air-gapped networks.

The CNAB spec is pretty open-ended about how to implement the spec, but Porter
took the route of making it incredibly easy to take existing tools, like
kubectl, Terraform, the azure/aws/gcloud CLIs, and use them inside a bundle. So
that you don't need to rewrite existing scripts.

## If an upgrade fails, can I roll back?

Bundles can be as smart as the bundle author and the logic that they put into
it! 😁 If an action fails, Porter logs that the upgrade was attempted and
failed, and will happily let you try again as many times as you want. Ideally
the bundle author planned ahead, and the logic they put into the upgrade action
is "re-runnable", or perhaps they provided custom actions to help remediate a
borked state.

One gap at the moment in the spec that we are still working on that makes the
above scenario more difficult is that there isn't a good spot in the spec for
the author to store state. For example, if during install they provisioned a VM
and have an instance ID they would like to store and use later, it's up to the
bundle author to find a spot to store that data right now. There isn't a
standard mechanism defined by the CNAB spec yet.

## How do I run commands that aren't in the default invocation image?

When you create a new bundle, porter generates Dockerfile.tmpl file for you. In
the porter.yaml you can specify `dockerfile: dockerfile.tmpl` to tell Porter
that you want to use the template (see [Custom Dockerfile](/bundle/custom-dockerfile/)) and then you can customize
it however you need. You're on the right track, from there you can use the exec
mixin to call whatever you installed.

We have been making custom mixins to provide a nice out-of-the-box user
experience. For example, last week I wrote a gcloud mixin that handles always
setting the `--format json` flag, parsing the json output and extracting the
values from the output and making them available to the rest of the bundle as
outputs that can be used in other steps or as output from the entire bundle.
That's something the exec mixin can't do (yet) and would be a reason to make a
custom mixin.

## Are mixins just wrappers around OS or executable calls?

Some Porter mixins are simple adapters between Porter and a command-line tool.
The Azure mixins are going to be full-fledged tools that communicate with the
Azure APIs, and provide an improved user experience. It all depends on how much
time you want to invest, and what you are starting from.

## Does Porter have its own registry?

No, Porter does not have a dedicated registry for hosting bundles. You can use any OCI registry, such as DockerHub, with Porter. We have compiled a [list of compatible registries](/compatible-registries/) that have been tested with Porter.

## What do mixins do?

Mixins provide the following features:
 * Install a tool into the bundle. It handles editing your dockerfile for you.
 * Mixins can adapt imperative command line tools to work with desired state. 
 
    For example, the uninstall command should be re-runnable. I should be able to re-run porter uninstall if there was an error and it should pick up where it left off. Most command line tools are imperative though, and don’t handle errors in a way that works well with that (e.g. returning a failure when you attempt to delete something that   is already deleted).
 * Mixins provide rich metadata that end-users can use to both understand what a bundle will do, and companies can limit what bundles they allow based on both the mixins present and what functionality of the mi   xin is used.
 * Mixins help collect outputs from steps, and reuse those as arguments to subsequent steps. 
    For example, the helm mixin can create a database and generate an output with the connection string. The terraform mixin can then use that connection string as an input variable, or even expose that output to the end-user when the bundle finishes running.
 
You never have to use mixins, other than the built-in exec. You could write a custom dockerfile and then call a bash script. But most people find that working with mixins is easier.

## What is the difference between Porter and existing DevOps tools?

Porter is never a replacement for an existing tool. Think of it as doing extra nice things on top of what those great tools already do!
 * Packages everything you use to deploy in a single artifact that can be easily distributed over registries and across air gapped networks.
 * Bundles can be signed and the signature verified before installing the bundle to improve supply chain security.
 * Makes for more reliable deployments because it has the exact version of a CLI that goes with the deployment. End users don’t need to install the right versions of tools locally (sometimes different apps require different versions).
 * Provides metadata about a deployment that you can use in many ways. End users can run porter explain to quickly see how to customize an installation with parameters, what credentials are needed by the bundle. No need to read separate installation docs to figure out how to deploy an app.
 * Reduces the operational knowledge required to manage an application. Often a deployment uses multiple tools, for example terraform, AND helm, AND kubectl and they are glued together with bash scripts. When those are put inside a bundle, it simplifies how much the end user needs to know. It’s always a single command, porter upgrade . In many cases the end user doesn’t even need to know what tools are used, or how to use those tools.
 * Makes it easier for a team to manage an application securely. Porter remembers the actions performed previously on the bundle, the parameter values used previously, the current version of the bundle, etc. 
 * You can put together the necessary parameters and credentials so your teammates don’t need to each hunt down the proper values themselves, or even worse copy sensitive credentials into local environment variables on every machine that does the deployment. Secrets stay in a secret store and aren’t copied around.
Those are just some of the reasons why working with bundles with the tools embedded inside is helpful vs using the same tools standalone.

## Do you have suggestions for bundles that I could make?

 * Create a bundle for a side project. It handles setting up any infrastructure, such as creating a VM or other cloud resources, and it then also deploys your app. Maybe pushing your serverless function as one example, or running a helm chart. 
   This is useful because often we figure out how to do these things while working on the project and then after not having time for the project most people forget how to deploy their project. That makes it super hard to start back up on the project! So using bundles helps make the best use of your limited side project time
* Create a bundle for an open source project. Pick a project that you like, maybe wordpress, mysql or discourse, and then write a bundle that will deploy the software to a cloud provider. This is a fun one because other people can use your bundle!

[CNAB Core Spec]: https://github.com/cnabio/cnab-spec/blob/main/100-CNAB.md
[CNAB Dependencies Spec]: https://github.com/cnabio/cnab-spec/blob/main/500-CNAB-dependencies.md
