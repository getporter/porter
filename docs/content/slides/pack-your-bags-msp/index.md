---
title: "Pack Your Bags: Managing Distributed Applications with CNAB"
description: |
  Learn how to use Cloud Native Application Bundles (CNAB) and Porter to bundle up your
  application and its infrastructure so that it is easier to manage.
url: "/pack-your-bags-msp/"
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

.center[👩🏽‍✈️ https://getporter.org/pack-your-bags-msp/#setup 👩🏽‍✈️ ]

1. Go to https://labs.play-with-docker.com/
1. Sign in with your Docker Hub account, or create one if you don't already have an account.
1. Click `Add new instance`
1. [Install Porter](/install)
   ```
   curl https://deislabs.blob.core.windows.net/porter/latest/install-linux.sh | bash
   export PATH=$PATH:~/.porter
   ```
1. Clone the workshop repository
    ```
    git clone https://github.com/getporter/porter.git
    cd porter/workshop
    ```

---
name: agenda

# Agenda

1. What is CNAB? - 15 minutes
1. See some bundles - 10 minutes
1. Build a bundle from scratch - 20 minutes
1. Walkthrough Porter - 10 minutes
1. The state of CNAB - 5 minutes
1. Choose your own adventure - 30 minutes

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
  <div class="right">
    <img src="/images/jerrycar.jpg" width="200px" />
    <p>Jeremy Rickard</p>
    <p>Senior Software Engineer</p>
    <p>Microsoft Azure</p>
  </div>
</div>

---
name: hi
exclude: true
# Stop and Say Hi

1. Move up to the front tables.
1. Introduce yourself to at least 2 people next to you.
2. Find out what they enjoy working on.
3. Share what talks or parts of Velocity they are looking forward to.

---
name: kickoff
class: center, middle
exclude: true
# First A Quick Demo!

---
# What is an application

---
name: cnab
# What's a CNAB?

---
name: use-cases
# When would you use a bundle?

1. Install the tools to manage your app: helm, aws/azure/gcloud, terraform
1. Deploy an app along with its infra: cloud storage, dns entry, load balancer, ssl cert
1. Get software and its dependencies into airgapped networks
1. Manage disparate operational tech: such as Helm, Chef, or Terraform, across teams and departments

---
class: center, middle
exclude: true
# Let's Answer That With A Story!

---
class: center, middle
exclude: true
# The Cast

---
class: center, middle
exclude: true
# You!
.center[
  ![you, a developer](/images/pack-your-bags/you-a-developer.jpg)
]

---
class: center, middle
exclude: true
# Your friend!
.center[
  ![your friend, a computer user](/images/pack-your-bags/your-friend-a-user.jpg)
]

---
class: center, middle
exclude: true
# Your friend!
.center[
  ![your friend, a computer user](/images/pack-your-bags/your-friend-a-user.jpg)
]

---
class: center, middle
exclude: true
# Your App!
.center[
  ![it's the journey that matters](/images/pack-your-bags/mcguffin.png)
]

---
class: center, middle
exclude: true
# Your Fans!
.center[
  ![trending](/images/pack-your-bags/your-fans.jpg)
]

---
class: center, middle
exclude: true
# Act One!

---
class: center, middle
exclude: true
# You Built an App

.center[
  ![you again](/images/pack-your-bags/you-a-developer.jpg)
  ![it's the journey that matters](/images/pack-your-bags/mcguffin.png)
]

---
class: center, middle
exclude: true
# It Runs Happily In The Cloud

---
exclude: true
# ....your cloud
.center[
  ![that's a bingo](/images/pack-your-bags/cloud-bingo.png)
]

---
class: center, middle
exclude: true
# Act Two!

---
class: center, middle
exclude: true
# Your Friend Wants To Run It!
.center[
  ![your friend, a computer user](/images/pack-your-bags/your-friend-a-user.jpg)
]

---
class: center, middle
exclude: true
# How exciting! 

---
exclude: true
# So you write extensive docs
.center[
  ![you fight for the users](/images/pack-your-bags/scroll-of-truth.png)
]

---
class: center, middle
exclude: true
# You are no longer friends...

.center[
  ![you fight for the users](/images/pack-your-bags/Spongebob-patrick-crying.jpg)
]

.footnote[http://vignette3.wikia.nocookie.net/spongebob/images/f/f0/Spongebob-patrick-crying.jpg/revision/latest?cb=20140713205315]

---
class: center, middle
exclude: true
# So you work together...

.center[
  ![pair programming](/images/pack-your-bags/working-together.jpg)
]

---
class: center, middle
exclude: true
# Then you help a few more people...

.center[
  ![go team](/images/pack-your-bags/go-team.jpg)
]

---
class: center, middle
exclude: true
# Act Three!

---
class: center, middle
exclude: true
# Suddenly McGuffin has FANS!

.center[
  ![all the github stars!!!](/images/pack-your-bags/your-fans.jpg)
]

---
class: center, middle
exclude: true
# This won't scale...

.center[
  ![nobody wants to do this](/images/pack-your-bags/scroll-of-truth.png)
]

---
class: center, middle
exclude: true
# So what do we do...

.center[
  ![this is my thinking face](/images/pack-your-bags/thinking.jpg)
]

---
class: center
# Containers ship our application

.center[
  ![ship it](/images/pack-your-bags/container-ship.jpg)
]

---
class: middle
# ... but don't install everything

<!--
.center[
  ![half way there](/images/pack-your-bags/scroll-of-sad-truth.png)
]
-->

---
# This is what CNAB solves

* Package All The Logic To Make Your App Happen
* Allow Consumer To Verify Everything They Will Install
* Distribute Them In Verifiable Way

---
class: center, middle
# How it works

.center[
  ![workflow](/images/pack-your-bags/the-workflow.png) ![magic](/images/pack-your-bags/magic.gif)
]

.footnote[_http://www.reactiongifs.com/magic-3_]

---
## Try it out: Install a bundle

```
$ porter install --reference deislabs/porter-hello-devopsdays:latest
```

---
name: anatomy
class: center, middle
# Anatomy of a Bundle

.center[
  ![so what is it](/images/pack-your-bags/anatomy.png)
]

---
# Application Images

* CNAB doesn't change this
* Build your application like you do now

---
# The Invocation Image

* MSI for the Cloud
* It's a Docker Image
* It contains all the tools you need to install your app
* It contains configuration, metadata, templates, etc

---
class: center, middle
# The Invocation Image

.center[
  ![so what is it](/images/pack-your-bags/easy-bake-oven-image.png)
]

---
# The Bundle Descriptor

* JSON!
* List of the invocation image(s) with digests
* List of the application image(s) with digests
* Definitions for inputs and outputs
* Can be signed

---

# CNAB Specification

* The Bundle format
* Defines how things are passed into and out of the invocation image
* A required entrypoint in invocation image
* Well-defined verbs
  * Install
  * Upgrade
  * Uninstall

---
# Are we done?

* We can install (complicated) things
* We can verify what we are going to install
* But how do we distribute bundles?

---
class: center
# Sharing Images With OCI Registries

![how docker shares](/images/pack-your-bags/ship-it.png)

---
name: registry
class: center, middle
# OCI Regristry ~ Docker Registry

---

# Distributing App and Invocation Images is solved
--

## So what about the bundle?
--

## It turns out OCI can help here too...

---

# OCI Registries Can Store Lots of Things

* CNAB today is working within the OCI Spec (not optimal)
* CNAB Spec group working with OCI to improve this

---
class: center

# Sharing Bundles With OCI Registries

![how oci shares bundles](/images/pack-your-bags/share-bundles.png)

---
class: center, middle
exclude: true
# An Example: Azure MySQL + Wordpress

.center[
  https://getporter.org/examples/src/azure-wordpress
]

---
class: center, middle

# Let's make some bundles!

.center[
  🚨 Not Setup Yet? 🚨

  https://getporter.org/pack-your-bags-msp/#setup
  
  ]

---
name: hello
class: center

# Tutorial
# Whale Hello

.center[
  ![whale saying hello](/images/whale-hello.png)
]

---

## porter create

```
$ porter create --help
Create a bundle. This generates a porter bundle in the current directory.
```

---

### porter.yaml
**The most important file**.  Edit and check this in. Everything else is optional.

### README.md
Explains the other files in detail

### Dockerfile.tmpl
Optional template for your bundle's invocation image

### .gitignore
Suggested set of files to ignore in git

### .dockerignore
Suggested set of files to not include in your bundle

---

# porter.yaml

```yaml
mixins:
  - exec

name: HELLO
version: 0.1.0
description: "An example Porter configuration"
tag: getporter/porter-hello

install:
  - exec:
      description: "Install Hello World"
      command: bash
      flags:
        c: echo Hello World
```

---

## Try it out: porter create

```
$ mkdir hello

$ cd hello

$ porter create
creating porter configuration in the current directory

$ ls
Dockerfile.tmpl  README.md  porter.yaml
```

---

## porter build

```
$ porter build --help
Builds the bundle in the current directory by generating a Dockerfile 
and a CNAB bundle.json, and then building the invocation image.
```

---

## Try it out: porter build

```
$ porter build

Copying dependencies ===>
Copying porter runtime ===>
Copying mixins ===>
Copying mixin exec ===>
Generating Dockerfile =======>
Writing Dockerfile =======>
Starting Invocation Image Build =======>
Generating Bundle File with Invocation Image porter-hello:latest =======>
Generating parameter definition porter-debug ====>
```

---

# What did Porter do? 🔎

---
### Dockerfile
```Dockerfile
FROM quay.io/deis/lightweight-docker-go:v0.2.0
FROM debian:stretch-slim
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ARG BUNDLE_DIR

COPY .cnab /cnab
COPY . ${BUNDLE_DIR}
RUN rm -fr ${BUNDLE_DIR}/.cnab
# exec mixin has no buildtime dependencies

WORKDIR ${BUNDLE_DIR}
CMD ["/cnab/app/run"]
```
.footnote[🚨 Generated by Porter]

---

### .cnab/
```
$ tree .cnab/
.cnab
├── app
│  ├── mixins
│  │  └── exec
│  │      ├── exec
│  │      └── exec-runtime
│  ├── porter-runtime
│  └── run
└── bundle.json
```

### .cnab/app/run

```bash
#!/usr/bin/env bash
exec /cnab/app/porter-runtime run -f /cnab/app/porter.yaml
```
.footnote[🚨 Generated by Porter]

---

### .cnab/bundle.json
```json
{
    "description": "An example Porter bundle",
    "invocationImages": [
        {
            "image": "deislabs/porter-hello-devopsdays-installer:latest",
            "imageType": "docker"
        }
    ],
    "name": "porter-hello-devopsdays",
    "parameters": {
        "fields": {
            "porter-debug": {
                "definition": "porter-debug",
                "description": "Print debug information from Porter when executing the bundle",
                "destination": {
                    "env": "PORTER_DEBUG"
                }
            }
        }
    },
    "definitions": { ...
```
.footnote[🚨 Generated by Porter]

---

## porter install

```
$ porter install --help
Install a bundle.

The first argument is the name of the claim to create for the installation. 
The claim name defaults to the name of the bundle.

Flags:
  -c, --credential-set strings         Credential to use when installing the bundle. 
  -f, --file string          Path to the bundle file to install.
      --param strings        Define an individual parameter in the form NAME=VALUE.
      --param-file strings   Path to a parameters definition file for the bundle
```

---
name: execution
## CNAB: What Executes Where

![cloud picture](/images/pack-your-bags/cnab-execution.png)

---

## Try it out: porter install

```
$ porter install

installing HELLO...
executing porter install configuration from /cnab/app/porter.yaml
Install Hello World
Hello World
execution completed successfully!
```

---
name: talkback

# Talk Back

* What tools do you use to deploy?
* Is it to a cloud? On-premise?
* Are you using a mix of tooling and platforms?
* What does your deployment lifecycle look like?

---
name: terraform
# A tale of two bundles

You can find these bundles in this repository under and **workshop/scratch/azure** and **workshop/porter-tf/azure**

---
# CNAB The Hardway

---
name: mellamo
class: center
# Tutorial
# Hi, My Name is _

.center[
  ![como se llamo me llamo llama](/images/me-llamo.jpg)
]

---
## Parameters

Variables in your bundle that you can specify when you execute the bundle
and are loaded into the bundle either as environment variables or files.

### Define a Paramer
```yaml
parameters:
- name: name
  type: string
  default: llama
```

### Use a Parameter
```yaml
- "echo Hello, ${ bundle.parameters.name }
```

* Needs double quotes around the yaml entry
* Needs double curly braces around the templating
* Uses the format `bundle.parameters.PARAMETER_NAME`

???
Explain defaults and when parameters are required

---
## Try it out: Print Your Name

Modify the hello bundle to print "Hello, YOUR NAME", for example "Hello, Aarti", using a parameter.

1. Edit the porter.yaml to define a parameter named `name`.
1. Use the parameter in the `install` action and echo your name.
1. Rebuild your bundle with `porter build`.
1. Finally run `porter install --param name=YOUR_NAME` and look for your name in the output.

---
### porter list

```
$ porter list
NAME          CREATED         MODIFIED        LAST ACTION   LAST STATUS
HELLO_LLAMA   5 seconds ago   3 seconds ago   install       success
HELLO         8 minutes ago   8 minutes ago   install       success
```

---
name: claims
### Claims

Claims are records of any actions performed by CNAB compliant tools on a bundle.

```
$ porter bundle show HELLO
Name: HELLO
Created: 8 minutes ago
Modified: 8 minutes ago
Last Action: install
Last Status: success
```

---
name: cleanup-hello
## Cleanup Hello World

First run `porter uninstall` without any arguments:
```
$ porter uninstall
uninstalling HELLO...
executing porter uninstall configuration from /cnab/app/porter.yaml
Uninstall Hello World
Goodbye World
execution completed successfully!
```

Now run `porter uninstall` with the name you used for the modified bundle:
```
$ porter uninstall HELLO_LLAMA
uninstalling HELLO_LLAMA...
executing porter uninstall configuration from /cnab/app/porter.yaml
Uninstall Hello llama
Goodbye llama
execution completed successfully!
```

---
# Publishing Bundles

Use `porter publish` to share bundles:

* Publishes both the invocation image and the CNAB bundle manifest to an OCI registry.
* Uses Docker tags for both

```yaml
name: HELLO-LLAMA
version: 0.1.0
description: "An example Porter configuration with moar llamas"
tag: "YOURNAME/porter-hello-llama"
```

---
# Modify the bundle

Let's make this bundle yours by changing it from the deislabs Docker Hub registry to your own:

1. Open **porter.yaml**
1. Find the `invocationImage` and the `tag` fields, and then change `deislabs` to your docker username.

    They should now look like this:

    ```yaml
    invocationImage: "YOURNAME/porter-hello-llama:latest"
    tag: "YOURNAME/porter-hello-llama-bundle:latest"
    ```
1. Save and close the file.

---
## porter publish

```
$ porter publish --help
Publishes a bundle by pushing the invocation image and bundle to a registry.


Examples:
  porter publish

```

---
# Try it out: Publish your bundle

```
$ porter publish
```

If you run into trouble here, here are a few things to check:
* Make sure that you updated the invocationImage and tag to use your docker hub username
* Log into docker hub with `docker login`

---
# Try it out: Install the bundle

```
$ porter install --reference YOURNAME/porter-hello-llama:v0.1.0 --param name=YOURNAME
```

---
class: center, middle
exclude: true
# BREAK

bring in the other slides that explain porter meta data, mixins, 
parameters and credentials

publish, install from our our tag before #4
_I_ may be comforable with terraform and _I_ know what credentials / parameters are needed
The person who writes the bundle isn't the person who runs the bundle

5 goes away



---
name: wordpress
class: center
exclude: true
# Tutorial
# Wordpress

---
exclude: true
## porter credentials generate

```
$ porter credentials generate --help
Generate a named set of credentials.

The first argument is the name of credential set you wish to generate. If not
provided, this will default to the bundle name. By default, Porter will
generate a credential set for the bundle in the current directory. You may also
specify a bundle with --file.

Bundles define 1 or more credential(s) that are required to interact with a
bundle. The bundle definition defines where the credential should be delivered
to the bundle, i.e. at /home/nonroot/.kube. A credential set, on the other hand,
represents the source data that you wish to use when interacting with the
bundle. These will typically be environment variables or files on your local
file system.

When you wish to install, upgrade or delete a bundle, Porter will use the
credential set to determine where to read the necessary information from and
will then provide it to the bundle in the correct location.
```

---
exclude: true
## Wordpress Credential Mapping

### ~/.porter/credentials/wordpress.yaml
```yaml
name: wordpress
credentials:
- name: kubeconfig
  source:
    path: /Users/carolynvs/.kube/config
```

### porter.yaml
```yaml
credentials:
- name: kubeconfig
  path: /home/nonroot/.kube/config
```

---
name: hack
exclude: true
### A quick hack

If you are using for Docker for Desktop with Kubernetes

1. Copy $HOME/.kube/config to $HOME/.kube/internal-config
1. Edit **internal-config** and change the server from `localhost` to `host.docker.internal`

```
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: https://host.docker.internal:6443
  name: docker-for-desktop-cluster
```

Specify this config file for `porter credentials generate` on the next slide for the kubeconfig.

---
exclude: true
## Try it out: porter credentials generate

Generate a set of credentials for the wordpress bundle in this repository.

1. Change to the `wordpress` directory under the workshop materials
1. Run `porter build`
1. Run `porter credentials generate` and follow the interactive prompts to create a set of credentials
for the wordpress bundle.

???
we all do this together

---
exclude: true
## Try it out: porter install --credential-set

Install the wordpress bundle and pass it the named set of credentials that you generated.

```
$ porter install --credential-set wordpress
```

---
name: cleanup-wordpress
exclude: true
## Cleanup Wordpress

```
$ porter uninstall --credential-set wordpress
```

???
Explain why --credential-set is required again for uninstall 

---
name: author
class: center, middle

# Authoring Bundles

---
name: manifest
# Porter Manifest In-Depth

---
## Metadata

```yaml
name: porter-workshop-tf 
version: 0.1.0
description: "An example using Porter to build the from scratch bundle"
tag: getporter/porter-workshop-tf
```

---
## Mixins

Declare any mixins that you are going to use

```yaml
mixins:
  - azure
  - terraform
```

---
name: parameters
## Parameters

Define parameters that the bundle requires.

```yaml
parameters:
  - name: location
    type: string
    default: "EastUS"

  - name: server-name
    type: string

  - name: database-name
    type: string
```

---
name: credentials
## Credentials

Define credentials that the bundle requires and where they should be placed
in the bundle when it is executing:

* environment variables (env)
* files (path)

```yaml
credentials:
  - name: subscription_id
    env: AZURE_SUBSCRIPTION_ID

  - name: tenant_id
    env: AZURE_TENANT_ID

  - name: client_id
    env: AZURE_CLIENT_ID

  - name: client_secret
    env: AZURE_CLIENT_SECRET
```

---
## Credentials

Variables that can be specified when the bundle is executed that are _associated with the identity 
of the user executing the bundle_, and are loaded into the bundle either as environment variables or files.

They are mapped from the local system using named credential sets, instead of specified on the command-line.

---
name: creds-v-params
## Credentials vs. Parameters

### Parameters
* Application Configuration
* Stored in the claim
* 🚨 Available in **plaintext** on the local filesystem

### Credentials
* Identity of the user executing the bundle
* Is not stored in the claim
* Has to be presented every time you perform an action

---
name: passwords
## Credentials, Passwords and Sensitive Data

* Credentials are for data identifying data associated with a user. They are 
re-specified every time you run a bundle, and are not stored in the claim.
* Parameters can store sensitive data using the `sensitive` flag. This prevents 
the value from being printed to the console.
* We (porter) and the CNAB spec are working on more robust storage mechanisms for 
claims with sensitive data, and better ways to pull data from secret stores so that 
they don't end up on the file system unencrypted.

In all honesty this area is a work in progress. I would shove anything sensitive in a 
credential for now but be aware of the distinction and where the CNAB spec is moving.

---
name: outputs
### Bundle Outputs

Often, a bundle will produce _something_. This might be a new certificate, a hostname, or some other piece of information. Bundle Outputs are used to provide this information back to the CNAB runtime.  

```
outputs:
  - name: STORAGE_ACCOUNT_KEY
    type: string
```

---
name: outputs
### View Bundle Outputs with Porter

```
$ porter bundle show


Outputs:
-------------------------------------------------------------------------------------------
  Name                 Type    Value
-------------------------------------------------------------------------------------------
  STORAGE_ACCOUNT_KEY  string  JKb9C+J+nFtGrDyBW4Y0zaIK5hzIvi2gW3SfnmnkcunyXSYV3HucQGNIo...
```

---
name: dockerfile
## Custom Dockerfile

Specify a custom Dockerfile for the invocation image

* Use a different base image
* Add users, tweak the environment and configuration
* Install tools and applications

⚠️ You are responsible for copying files into the bundle into `${BUNDLE_DIR}`

---

### porter.yaml

```yaml
dockerfile: Dockerfile.tmpl
```

### Dockerfile.tmpl

```Dockerfile
FROM debian:stretch-slim

ARG BUNDLE_DIR

RUN apt-get install -y curl

COPY myscript.sh ${BUNDLE_DIR}
```

---
exclude: true
## Try it out: Custom Dockerfile

Make a bundle that uses a custom `dockerfile` template and uses **mcr.microsoft.com/azure-cli**
as its base image

1. Create a porter bundle in a new directory with `porter create`.
1. Modify the **porter.yaml** to uncomment out `#dockerfile: Dockerfile.tmpl`.
1. Edit **Dockerfile.tmpl** to use **mcr.microsoft.com/azure-cli** for the base image.
1. Edit the install action to run `az help`.
1. Build the bundle.
1. Install the bundle.

---
## Porter's Default Dockerfile

* Uses Debian for the base image
* Installs root ssl certificates

🎩✨ Automatically copies everything into the bundle for you

---
# Actions and Steps

---
## Actions

Actions map to the verbs you use when you use Porter.

* porter install
* porter upgrade
* porter uninstall

These are defined in the CNAB specification. You can also define your own actions and run them

```
$ porter invoke --action myaction
```

---
## Steps

Within an action you can define a series of ordered steps.

* A step must complete successfully before the next step is executed.
* A step is defined using a mixin. Each step can use only one mixin. So far we have
used the `exec` mixin that lets you run commands.
* Mixins must be declared ahead of time in the `mixins` section of the manifest.

---

```yaml
install: # action
  - exec: # step
      description: "Install my application"
      command: bash
      arguments:
        - install-myapp.sh
```

---
### Referencing files from the Manifest

We recommend referencing files using relative paths

```
$ tree
├── Dockerfile
├── README.md
├── porter.yaml
└── scripts
*    └── do-things.sh
```

**porter.yaml**
```yaml
install:
  - exec:
      description: Do Things
      command: bash
      arguments:
*        - scripts/do-things.sh
```

---
## Step Outputs

* You tell the mixin what data to extract and put into an ouput
* Each mixin output will look different, but they all require a `name`
* Porter stores the output value, making it available to later steps
* Not all mixins support outputs

---
exclude: true
### Helm Mixin Output

```yaml
install:
- helm3:
    ...
    outputs:
    - name: mysql-password
      secret: mydb
      key: mysql-password
```

---
exclude: true
### Kubernetes Mixin Output

```yaml
install:
  - kubernetes:
      description: "Create NGINX Deployment"
      manifests:
        - manifests/nginx
      wait: true
      outputs:
        - name: "IP_ADDRESS"
          resourceType: service
          resourceName: nginx-deployment
          jsonPath: "{.spec.clusterIP}"
```

---
name: wiring

## Hot Wiring the Manifest

These variables are available for you to use in the manifest:

* bundle.name
* bundle.parameters.PARAMETER_NAME
* bundle.credentials.CREDENTIAL_NAME
* bundle.outputs.OUTPUT_NAME

---
name: templating

## Templating

Porter uses a template engine to substitute values into the manifest.

* Needs double quotes around the yaml entry
* Use double curly braces around the templated value

**Example**
```yaml
connectionString: ${bundle.outputs.host}:${bundle.outputs.port}
```

---
name: mixins
class: center, middle

# Mixins

_"Mixins are the lifeblood of Porter._

_They adapt between CNAB and existing tools. Porter is just glue."_

---
## Mixins Available Today

* exec
* kubernetes
* helm
* azure
* aws
* gcloud
* terraform

.center[ https://getporter.org/mixins ]

---
name: helm
background-image: url(/images/pack-your-bags/helm-mixin.png)

---
exclude: true
## Try it out: Install a Helm Chart

Make a new bundle and install the Helm chart for etcd-operator

1. Create a porter bundle in a new directory with `porter create`.
1. Modify the **porter.yaml** to use the **helm** mixin and define credentials for **kubeconfig**.
1. Using the helm mixin, install the latest **stable/etcd-operator** chart with the default values.
1. Build the bundle.
1. Generate credentials for your bundle.
1. Install your bundle.

---
## Installing Mixins

Anyone can make a mixin and have people install it using Porter

**porter mixin install terraform**

---

## What if I need a mixin that doesn't exist? 😰

You can always use a custom dockerfile to install the tool you need and then
execute the commands with the exec mixin.

Custom mixins are just easier to use, but aren't necessary.

---
exclude: true
# Make Your Own Mixin

---
exclude: true
## What makes a mixin?

* Executable written in any language
* Communicates to Porter on stdin and stdout
* Supports a few commands: build, schema, install, upgrade, uninstall
* Translates the steps from the porter manifest to commands against any external tool or service

---
exclude: true
## What mixin would you make?

* Docker
* Google Cloud
* CloudFormation
* AWS
* Artifactory
* Vault
* Dominos 🍕
* What else?

---
exclude: true
# Now run your bundle

* Run `porter install --reference [your new tag]`

Example tag of `cnabaholic/hello-people:latest`:

* Run `porter install --reference cnabaholic/hello-people:latest`

---
exclude: true
# How does it do that?

* Updates the bundle with the new images and digests
* Stores the bundle information as a combination:
  * OCI manifest list annotations to the main manifest list for the tag
  * A new manifest list for the bundle credentials and parameters.
* When pulling a bundle, it reconstructs it from the parts mentioned above

See [OCI Bundle Format](/oci-bundle-format) for an example.

---
name: survey
# CNAB Tooling and SDKs

* Porter
* Docker App
* Duffle
* cnab-go

https://cnab.io/community-projects

---
# State of CNAB

* CNAB Core Spec is release 1.0 _very soon_
* CNAB Signing Spec is underway
* Tooling is finishing support for CNAB Core

---
# Questions and Feedback

???
What should someone do if they are interested in CNAB for their work or personal projects?
What is the timeline for the project and how should they be thinking about beginning to incorporate it?

---
# Choose your own adventure!

* ASCII Art Gophers - workshop/asciiart
* Create a MySQL on Azure - workshop/porter-tf-aci
* Manage VMs with gcloud - workshop/gcloud-compute
* Manage Buckets with aws - workshop/aws-bucket
