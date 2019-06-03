---
title: "Pack Your Bags: Managing Distributed Applications with CNAB"
description: |
  Learn how to use Cloud Native Application Bundles (CNAB) and Porter to bundle up your
  application and its infrastructure so that it is easier to manage.
url: "/pack-your-bags/"
---
class: center, middle

# Pack your bags
##  Managing Distributed Applications with CNAB

---
name: setup

# Workshop Setup

It can take a while for things to download and install over the workshop wifi,
so please go to the workshop materials directory and follow the setup instructions
to get all the materials ready.

.center[👩🏽‍✈️ https://porter.sh/pack-your-bags/#setup 👩🏽‍✈️ ]

* Clone the workshop repository
  ```
  git clone https://github.com/deislabs/porter.git
  cd porter/workshop
  ```
* [Install Porter](https://porter.sh/install)
* Create a Kubernetes Cluster on [macOS](https://docs.docker.com/docker-for-mac/kubernetes/) or [Windows](https://docs.docker.com/docker-for-windows/kubernetes/)
* [Install Helm 2](https://helm.sh/docs/install/)
* Initialize Helm on your cluster by running `helm init`


---
name: agenda

# Agenda

1. What is CNAB?
2. Manage Bundles with Porter
3. Authoring Bundles

---
name: introductions

# Introductions

😅<br/>
Carolyn Van Slyck<br/>
Senior Software Engineer<br/>
Microsoft Azure<br/>

😎<br/>
Jeremy Rickard<br/>
Senior Software Engineer<br/>
Microsoft Azure<br/>

---
name: cnab

# What is CNAB?

???
Explain why a spec is necessary

---
name: anatomy

# Anatomy of a Bundle

---
name: sharing

# Sharing Bundles

---

# Demo

## porter push + docker app install

---
class: center, middle

# BREAK

---
class: center, middle

# Manage Bundles with Porter

.center[
  🚨 Not Setup Yet? 🚨

  https://porter.sh/pack-your-bags/#setup
  
  ]
---

# Hello World Tutorial

---

## porter create

```console
$ porter create --help
Create a bundle. This generates an empty porter bundle for you to customize.
```

---

### porter.yaml


### README.md

### Dockerfile.tmpl

---

## Try it out

```console
$ mkdir hello
$ porter create
creating porter configuration in the current directory
$ ls
Dockerfile.tmpl  README.md  porter.yaml
```

---
# Hello X Tutorial 

---

# Wordpress Tutorial

---
class: center, middle

# BREAK

---
class: center, middle

# Authoring Bundles

---

# Porter Manifest In-Depth

---

# Steps and Actions

---

# Wiring

---

# Templating

---
class: center, middle

# Mixins

---

# Step Outputs

---

# Make Your Own Mixin

---

# Break Glass

---
class: center, middle

# CNAB Best Practices

---

# What would you really put into a bundle?

---

# What does a real bundle look like?

???
Look at the azure examples and quick starts

---

# How does this fit into a CI/C pipeline?

---
class: center, middle

# Tooling

---

# CNAB Tooling Ecosystem

???
Explain where porter shines, what it is good at vs. say docker app

---

# Duffle

???
Mention duffle as a ops tool for managing bundles from multiple orgins at runtime

---
class: center, middle

# Beyond!

---

# Roadmap

???

Both CNAB and Porter for the next 3 months and rest of the year

---

# Next Steps

???
What should someone do if they are interested in CNAB for their work or personal projects?
What is the timeline for the project and how should they be thinking about beginning to incorporate it?

---

# Contribute!

---
class: center, middle

# Choose your own adventure

* Cloud + Break Glass
* Order a pizza with Porter
* Make a mixin