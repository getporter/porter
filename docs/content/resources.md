---
title: Resources
description: External videos, blog posts and tutorials about CNAB and Porter
---

Do you have a blog post, video, tutorial, demo, or some other neat thing 
using Porter or CNAB that you'd like to share? [Open up a pull request][pr] 
and show it off! ✨

* [Building a Terraform Based Bundle with Porter](#building-a-terraform-based-bundle-with-porter)
* [Porter Bundle with K3D, Helm3 and Brigade by Nuno Do Carmo](#porter-bundle-with-k3d-helm3-and-brigade-by-nuno-do-carmo)
* [Porter: An Opinionated CNAB Authoring Experience](#porter-an-opinionated-cnab-authoring-experience)
* [Free Glue Code - Porter](#free-glue-code-porter)

[pr]: https://github.com/deislabs/porter/blob/master/CONTRIBUTING.md

### Building a Terraform Based Bundle with Porter

This video shows how to use Porter to build Terraform based bundles. It also
shows how multiple technologies can be combined into a single bundle, in this
case both ARM and Terraform.

<iframe width="560" height="315" src="https://www.youtube.com/embed/LxRvKg3egPc" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

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