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

<h3 style="margin-top: 3em">
  <a href="https://porter.sh/cnab-unpacked/">porter.sh/cnab-unpacked</a>
</h3>

---
name: introductions
# Introductions

<div id="introductions">
  <div class="left">
    <img src="/images/carolynvs.jpg" width="150px" style="margin-top: 3em"/>
    <p style="margin-top:1em; font-weight: bold;">Carolyn Van Slyck</p>
    <p><a href="https://twitter.com/carolynvs">@carolynvs</a></p>
    <p>Senior Software Engineer</p>
    <p>Microsoft</p>
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

If I gave this to a friend to deploy, would they...
--

* Clone a repository? The app's or a devops one?
--

* Install specific versions of terraform and helm?
--

* Set environment variables, and save config files to specific locations?
--

* Use specific helm and terraform commands?
--

* Use a utility docker container that required them to mount volumes from the 
  local host and pass through environment variables?
--

* Guess all of this correctly... the first time? 😅
--

* How about at 2am while on-call for an app they didn't write? 😨
--

* Still be your friend? 🤔


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

Credentials
-------------------------------------------------------------------
| Name        | Type   | Description        |                      |
------------------------------------------------------------------- 
  kubeconfig    string   Path to kubeconfig  

Parameters
-------------------------------------------------------------------- 
| Name          | Type         | Description   | Default (Required) |  
-------------------------------------------------------------------- 
  sparkles        boolean       Moar ✨          false

```

🚧 https://github.com/deislabs/porter/issues/635

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
# Let's Reflect
--

* Self describing, so it can tell you what you need to install it
--

* Installed with a single command
--

* Underlying toolsets and logic were abstracted in the bundle
--

* Distributed via OCI (Docker) registry
--

* You are still friends 😎

---
# What was in the bundle?

The application **and everything needed to install it**

* Helm and terraform CLIs
* Helm chart
* Terraform files
* Bash script that orchestrates installing everything

---

# Awkward Question Time!
--

## 🙋🏻‍♀️ Does this replace &lt; my favorite tech &gt;?
--

## 🙋🏻‍♀️ Why wouldn't I just use &lt; my favorite tech &gt;?
--

## 🙋🏻‍♀️ I don't like the sound of that bash script...

---
class: middle
name: use-cases
# When would you use a bundle?

---
# Include required tools

## Distribute files in the CNAB invocation image

.center[
  ![so what is it](/images/pack-your-bags/easy-bake-oven-image.png)
]

---

# Deploy App's Infrastructure

## Custom script for the invocation image entrypoint

.center[
  <img src="/images/porter-mixin-cloud.png" alt="helm, terraform, gcloud, azure logo cloud" width="400px" />
]

---

# Airgapped Networks or Offline

## Thick bundles include referenced images

.center[
  <img src="/images/cnab-unpacked/usb-stick-cnab.svg" width="300px" />
]

---

# Manage multiple tech stacks

## Consistent interface regardless horrors inside

.center[
  <img src="/images/cnab-unpacked/how-to-make-a-bundle.gif" alt="man packing suitcase with his foot" width="300px"/>
]

---

# Immutable, verified installer

## Signed bundles referencing image digests

.center[
  <img src="/images/cnab-unpacked/signed-bundle.svg" width="300px" />
]

---
# CNAB Sub Specifications

## Core
## Registries 🚧
## Security 🚧
## Claims 🚧
## Dependencies 🚧

---
# Core Specification

* Bundle file format (bundle.json)
* Invocation image format, aka "the installer"'
* Entrypoint in invocation images
* Bundle format (thin or thick)
* Bundle runtime execution behavior
* Well-known Actions
  * Install
  * Upgrade
  * Uninstall
* Custom Actions

.center[
  **Version 1.0 was released this month!** 🎉
]

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

.center[
  **The Installer** or **MSI for the Cloud**
]

* Includes all the tools you need to install your app
* Has your configuration, metadata, templates, etc
* Run script with your logic for install, upgrade and uninstall

.center[
  <img src="/images/pack-your-bags/easy-bake-oven-image.png" width="275px" />
]

---
# The Bundle Descriptor

* **bundle.json**
* Invocation and Application images with their content digests
* Credentials and Parameters accepted by the installer
* Outputs generated by the installer

---
# Registries Specification

Push and pull bundles to OCI registries

.center[
  <img src="/images/pack-your-bags/share-bundles.png" alt="how oci shares bundles" width="600px" style="margin-top: 3rem;"/>
]

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

🚧 Very early stage 

* Require other bundles
* Specify their version
* Use their outputs

---
# CNAB Tooling

* Porter
* Docker App
* Duffle

Anyone can write their own too! These are all based on:

https://github.com/deislabs/cnab-go
--

## 🙋🏻‍♀️ Are all CNAB tools interchangeable?

---
# Porter

The friendly cloud installer that bootstraps your bundles using tools and assets from your current pipeline. ✨

* Doesn't require knowledge of CNAB
* Uses mixins to include tools into bundles
* Designed to make bundles easier to manage
* Community focused

.center[
  <img src="/images/porter-notext.png" alt="woman in black suit and hat, with cnab logo" width="300px" style="bottom: 0; position: absolute;"/>
]

---
# Demo

## Deploy a bundle with Porter

.nudge[.center[
  https://github.com/jeremyrickard/do-porter
]]

---
# Parting Awkward Questions
--

## 🙋🏻‍♀️ Are bundles ready to use?
--

## 🙋🏻‍♀️ Is Porter a Microsoft-only tool?
--

## 🙋🏻‍♀️ This is more of a comment really...
--

## 🙋🏻‍♀️ Ask me yours!

---
# Resources

* [cnab.io][cnab]
* [cnab.io/community-meetings/#communications][cnab-slack] - #cnab CNCF Slack
* [porter.sh][porter]
* [porter.sh/contribute][contribute] - New Contributor Guide
* [porter.sh/community][porter-slack] - #porter CNCF Slack

[cnab]: https://cnab.io
[cnab-slack]: https://cnab.io/community-meetings/#communications
[porter]: https://porter.sh
[contribute]: https://porter.sh/contribute
[porter-slack]: https://porter.sh/community