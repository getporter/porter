---
title: "CNAB Unpacked"
description: |
  Learn what is a Cloud Native Application Bundle and when you would use one.
url: "/cnab-unpacked/"
---
class: center, middle

## Understanding 
## Cloud Native Application Bundles
# CNAB Unpacked

---
name: introductions
# Introductions

<div id="introductions">
  <div class="left">
    <img src="/images/carolynvs.jpg" width="150px" />
    <p>Carolyn Van Slyck</p>
    <p>Senior Software Engineer</p>
    <p>Microsoft Azure</p>
  </div>
</div>

---
# What is CNAB?

.nudge[
> Cloud Native Application Bundles is an open-source packaging and distribution specification for managing distributed applications with a single installable file.
]

---
# Where did it come from?

.nudge[
<img float="left" src="/images/logos/azure.png" />
<img float="left" src="/images/logos/docker.png"/>
<img float="left" src="/images/logos/pivotal.png"/>
]

### We had contributions from other companies and the community! ❤️

---
# What does CNAB solve?

.nudge[
### The gap between your application's code and everything 
### necessary to deploy your application.
]

---
# Let's define an app

* Terraform to create the infrastructure
* Helm to deploy to a kubernetes cluster
* Obligatory bash script

---
# Let's find the gap

If I gave this to a customer, IT, friend or enemy to deploy, would they...
--

* Clone a repository? The app's or a special devops one?
--

* Install specific versions of terraform and helm?
--

* Set environment variables, and save files to special locations?
--

* Use the right helm and terraform commands?
--

* Use a utility docker container that required them to mount volumes from the 
  local host and pass through environment variables?
--

* Guess all of this correctly... the first time? 😅
--

* How about at 2am while on-call for an app they didn't write? 😨
--

* Still be your friend? 🤔
--

---
class: middle
# Let's try this with a bundle

---
# Get ready...

```
$ porter explain --tag deislabs/tron:v1.0
name: Tron
description: The classic game of light cycles and disc wars
version: 1.0.0

Parameters
-------------------------------------------------------------------- 
| Name          | Type         | Description   | Default (Required) |  
-------------------------------------------------------------------- 
  sparkles        boolean       Moar ✨          false

Credentials
-------------------------------------------------------------------
| Name        | Type   | Description        |                      |
------------------------------------------------------------------- 
  kubeconfig    string   Path to kubeconfig  

```

---
# Get Set...

```
$ porter credentials generate -t deislabs/tron:v1.0

Generating new credential azure from bundle tron
==> 1 credentials required for bundle tron
? How would you like to set credential "kubeconfig" file path
? Enter the path that will be used to set credential "kubeconfig"

Saving credential to /Users/carolynvs/.porter/credentials/azure.yaml
```

# Go!
```
$ porter install tron -t deislabs/tron:v1.0 --creds azure --param sparkles=true
```

---
## So what was in that bundle?

The application **and everything needed to install it**

* Helm and terraform CLIs
* Bash script that orchestrates installing everything
* Helm chart
* Terraform files


---
class: middle
name: use-cases
# When would you use a bundle?

---
# Include required tools

.nudge[

  ## Distribute files in the CNAB invocation image

]

---

# Deploy App's Infrastructure

.nudge[

  ## Custom script for the invocation image entrypoint

]

---

# Airgapped Networks or Offline

.nudge[

  ## Thick bundles include referenced images

]

---

# Manage multiple tech stacks

.nudge[

  ## Consistent interface regardless horrors inside

]

---

# Immutable, verified installer

.nudge[

  ## Signed bundles referencing image digests

]

---
# CNAB Sub Specifications

## Core
## Registries 🚧
## Security 🚧
## Claims 🚧
## Dependencies 🚧

🚧 in-progress

---
# Core Specification

* Bundle file format (bundle.json)
* Invocation image format, aka "the installer"'
* Entrypoint in invocation images
* Bundle format (thin or thick)
* Bundle runtime execution behavior
* Well-defined verbs
  * Install
  * Upgrade
  * Uninstall
* Custom Actions

Version 1.0 was released this month! 🎉

---
name: anatomy
class: center, middle
# Anatomy of a Bundle

.center[
  ![so what is it](/images/pack-your-bags/anatomy.png)
]

---
# Application Images

* The same same docker images you use now
* Continue to build and distribute them without change
* CNAB doesn't affect this

---
# The Invocation Image

* a.k.a The Installer or "MSI for the Cloud"
* Includes all the tools you need to install your app
* Has your configuration, metadata, templates, etc
* Run script with your logic for install, upgrade and uninstall

---
class: center, middle
# The Invocation Image

.center[
  ![so what is it](/images/pack-your-bags/easy-bake-oven-image.png)
]

---
# The Bundle Descriptor

* **bundle.json**
* Invocation and Application images with their content digests
* Credentials and Parameters accepted by the installer
* Outputs generated by the installer

---
class: center, middle
# How it works

.center[
  ![workflow](/images/pack-your-bags/the-workflow.png) ![magic](/images/pack-your-bags/magic.gif)
]

.footnote[_http://www.reactiongifs.com/magic-3_]

---
# Registries Specification

Push and pull bundles to OCI (docker) registries

![how oci shares bundles](/images/pack-your-bags/share-bundles.png)

---
# Security Specification

* Image digests
* Signing bundles
* Bundle attestation

---
# Claims Specification

Record actions performed on a bundle:

* Parameters passed
* Outputs generated
* Success/Failure

---
# Dependencies Specification

Very early stage.

* Require other bundles
* Specify their version
* Use their outputs

---
# CNAB Tooling

* Duffle
* Porter
* Docker App

---
# Demo

---
# 

* Does this replace existing technology?
* Why wouldn't I just use helm, terraform, etc?
* Are all CNAB tools the same? Interchangeable?