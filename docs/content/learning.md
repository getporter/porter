---
title: Learning
description: External videos, blog posts and tutorials about CNAB and Porter
aliases:
- /resources/
---

Do you have a blog post, video, tutorial, demo, or some other neat thing 
using Porter or CNAB that you'd like to share? [Open up a pull request][pr] 
and show it off! ✨

Check out our new [CNCF Porter Community] channel for all of our conference talks, demos and meeting videos.

* [The Devil is in the Deployments: Bundle Use Cases](#the-devil-is-in-the-deployments-bundle-use-cases)
* [Deploy Across an Airgap](#deploy-across-an-airgap)
* [Understanding Cloud Native Application Bundles](#understanding-cloud-native-application-bundles)
* [Porter: Digital Ocean, Terraform, Kubernetes](#porter-digital-ocean-terraform-kubernetes)
* [Porter Bundle with K3D, Helm3 and Brigade by Nuno Do Carmo](#porter-bundle-with-k3d-helm3-and-brigade-by-nuno-do-carmo)
* [Porter: An Opinionated CNAB Authoring Experience](#porter-an-opinionated-cnab-authoring-experience)
* [Free Glue Code - Porter](#free-glue-code-porter)

[pr]: /contribute/guide/

### The Devil is in the Deployments: Bundle Use Cases

Can you deploy your entire app from scratch with a Helm install? Or do you
have cloud infra and hosted services that you rely on? The cloudy bits that make
your app cloud native.

Cloud Native Application Bundles, the CNAB spec, was designed to solve
deployment problems that we all have been quietly battling with, mostly with
hope and bash. Bundles come in handy when deploying applications that don't live
neatly inside of just Kubernetes.

Let's learn when bundles make sense, when they don't, and what your day could
look like if you were using them:

* Install tools to manage your app: helm, aws/azure/gcloud, terraform.
* Deploy your app along with its infra: cloud storage, dns entry, load balancer, ssl cert.
* Get software and its dependencies into airgapped networks.
* Manage disparate operational tech, such as Helm or Terraform, across
  teams and departments.
* Secure your pipeline.

<iframe width="560" height="315" src="https://www.youtube.com/embed/wNl8m3h9I4E" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

### Deploy Across an Airgap

A gentle introduction to Porter and a demo of how to use Porter to deploy your application across an airgap.

<iframe width="560" height="315" src="https://www.youtube.com/embed/IFWIBSzhgM4" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

### Understanding Cloud Native Application Bundles

CNCF Webinar where Carolyn Van Slyck, the co-creator of Porter, the friendly cloud installer that gives you building blocks to create CNAB bundles from your existing pipeline, will demonstrate real world bundles, answering common questions:

* What is a bundle?
* When are bundles useful?
* Does this replace existing technology?
* Why wouldn’t I just use helm, terraform, etc?
* Are all CNAB tools the same? Interchangeable?

<iframe width="560" height="315" src="https://www.youtube.com/embed/1FGMrv_xfqY" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

### Porter: Digital Ocean, Terraform, Kubernetes

A demo video of using Porter to deploy infrastructure to Digital Ocean and Kubernetes using Terraform and Helm.

<iframe width="560" height="315" src="https://www.youtube.com/embed/ciA1YuGOIo4" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

### Porter Bundle with K3D, Helm3 and Brigade by Nuno Do Carmo

Nuno Do Carmo demonstrates how to use Porter and CNAB to install Brigade on a new Kubernetes cluster using k3d and Helm.

<iframe width="560" height="315" src="https://www.youtube.com/embed/9egipQjUgD0" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

### Porter: An Opinionated CNAB Authoring Experience

When we deploy to the cloud, most of us aren't dealing with just a single cloud
provider or even deployment tool. It seems like even the simplest of
applications today need load balancers, persistent storage, databases, SSL
certificates, and any number of other components. That's even before you get to
your own application! That is a lot to figure out!

The Cloud Native Application Bundle specification was created to help solve this
problem. CNAB is specifies lots of great things about a bundle and how it is
run, but it actually gives you a good deal of freedom in how to actually build
that bundle. Porter, a cloud native package manager built on CNAB, adopts an
opinionated approach to make bundle authoring straightforward and approachable.

In this talk, you will learn how Porter makes it easier to author CNAB bundles
and manage cloud native applications in the messy imperfect hybrid cloud world
that we all live in. You'll also learn some of the issues we encountered that
led us to develop Porter, how we addressed them, how you can contribute to
Porter's development and how you might build your own tooling on top of the CNAB
spec.

<iframe width="560" height="315" src="https://www.youtube.com/embed/__fim6RIW1s" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

### Free Glue Code - Porter

What problem does CNAB solve and where does Porter fit in? Porter maintainer, Carolyn Van Slyck does her best to explain
assisted as always with emoji and markdown.

<p align=center><a href="https://carolynvanslyck.com/blog/2019/04/porter">Free Glue Code - Porter</a></p>

[CNCF Porter Community]: /videos
