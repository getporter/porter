---
title: FAQ
description: Frequently Asked Questions
---

* [What is CNAB?](#what-is-cnab)
* [Does Porter fully implement the CNAB specification?](#does-porter-fully-implement-the-cnab-specification)
* [Does Porter solve something that Ansible, Terraform, etc does not?](#does-porter-solve-something-that-ansible-terraform-etc-does-not)
* [Does Porter Replace Duffle?](#does-porter-replace-duffle)
* [Should I use Porter or Duffle?](#should-i-use-porter-or-duffle)
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

1. Requesting credentials for an existing Kuberentes cluster.
1. Provisioning an Azure MySQL database.
1. Collecting the database credentials.
1. Installing the Wordpress chart and passing in the database credentials.

The person managing the application would only need to know `porter install` and have
credentials for a Kubernetes cluster.

## Does Porter fully implement the CNAB specification?

Porter currently implements much of the CNAB spec, however, as the [CNAB
specification](https://github.com/cnabio/cnab-spec) moves toward 1.0, some
gaps have emerged. Currently, if you build a bundle with Porter, you'll be able
to install it with Porter. There are some gaps with the spec that limit
compatibility with other CNAB tooling. See the [CNAB 1.0
Milestone](https://github.com/deislabs/porter/milestone/12) for more information
on these gaps.

## Does Porter solve something that Ansible, Terraform, etc does not?

Porter is an implementation of the CNAB specification. Cloud Native Application
Bundles is a different way of answering "_How do I reliably, securely deploy an
application and its infrastructure?_". It isn't replacing Ansible or Terraform
but adding some concepts on top. For example, packaging together the Terraform
binary and your Terraform scripts into an immutable bundle with a digest that
attests that the contents haven't been altered, that can be distributed via OCI
(docker) registries or USB sticks to get into air-gapped networks.

The CNAB spec is pretty open-ended about how to implement the spec, but Porter
took the route of making it incredibly easy to take existing tools, like
kubectl, Terraform, the azure/aws/gcloud CLIs, and use them inside a bundle. So
that you don't need to rewrite existing scripts.

## Does Porter replace Duffle?

  <p align="center"><strong>No, Porter is not a replacement of Duffle.</strong></p>

In short:

> Duffle is the reference implementation of the CNAB specification and is used 
> to quickly vet and demonstrate a working specification.

> Porter supports the CNAB spec and empowers bundle authors to create composable, 
> reusable bundles using familiar tools like Helm, Terraform, and their cloud provider's 
> CLIs. Porter is designed to be the best user experience for working with bundles.

See [Porter or Duffle](/porter-or-duffle) for a comparison of the tools.

## Should I use Porter or Duffle?

If you are contributing to the CNAB specification, we recommend vetting your contributions by
"verification through implementation" on Duffle.

If you are making bundles, may we suggest using Porter?

<p align="center">👩🏽‍✈️ ️️👩🏽‍✈️ 👩🏽‍✈️</p> 

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
that you want to use the template (see [Custom Dockerfile](/custom-dockerfile/)) and then you can customize
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
The Azure mixins are going to be full fledged tools that communicate with the
Azure APIs, and provide an improved user experience. It all depends on how much
time you want to invest, and what you are starting from.