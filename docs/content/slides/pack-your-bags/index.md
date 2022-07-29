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

.center[👩🏽‍✈️ https://getporter.org/pack-your-bags/#setup 👩🏽‍✈️ ]

* Clone the workshop repository
  ```console
  git clone https://github.com/getporter/porter.git
  cd porter/workshop
  ```
* [Install Porter](/install)
* Create a [Docker Hub](https://hub.docker.com/signup) account if you don't have one
* Create a Kubernetes Cluster on [macOS](https://docs.docker.com/docker-for-mac/kubernetes/) or [Windows](https://docs.docker.com/docker-for-windows/kubernetes/)


---
name: agenda

# Agenda

1. What is CNAB?
2. Manage Bundles with Porter
3. Authoring Bundles

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
# Stop and Say Hi

1. Move up to the front tables.
1. Introduce yourself to at least 2 people next to you.
2. Find out what they enjoy working on.
3. Share what talks or parts of Velocity they are looking forward to.

---
name: kickoff
class: center, middle
# First A Quick Demo!

---
# What is an application

---
name: cnab
# What's a CNAB???

---
class: center, middle

# Let's Answer That With A Story!

---
class: center, middle

# The Cast

---
class: center, middle

# You!
.center[
  ![you, a developer](/images/pack-your-bags/you-a-developer.jpg)
]

---
class: center, middle

# Your friend!
.center[
  ![your friend, a computer user](/images/pack-your-bags/your-friend-a-user.jpg)
]

---
class: center, middle

# Your friend!
.center[
  ![your friend, a computer user](/images/pack-your-bags/your-friend-a-user.jpg)
]

---
class: center, middle

# Your App!
.center[
  ![it's the journey that matters](/images/pack-your-bags/mcguffin.png)
]

---
class: center, middle

# Your Fans!
.center[
  ![trending](/images/pack-your-bags/your-fans.jpg)
]

---
class: center, middle

# Act One!

---
class: center, middle

# You Built an App

.center[
  ![you again](/images/pack-your-bags/you-a-developer.jpg)
  ![it's the journey that matters](/images/pack-your-bags/mcguffin.png)
]

---
class: center, middle

# It Runs Happily In The Cloud

---
# ....your cloud
.center[
  ![that's a bingo](/images/pack-your-bags/cloud-bingo.png)
]

---
class: center, middle

# Act Two!

---
class: center, middle

# Your Friend Wants To Run It!
.center[
  ![your friend, a computer user](/images/pack-your-bags/your-friend-a-user.jpg)
]

---
class: center, middle

# How exciting! 

--

# So you write extensive docs
.center[
  ![you fight for the users](/images/pack-your-bags/scroll-of-truth.png)
]

---
class: center, middle

# You are no longer friends...

.center[
  ![you fight for the users](/images/pack-your-bags/Spongebob-patrick-crying.jpg)
]

.footnote[http://vignette3.wikia.nocookie.net/spongebob/images/f/f0/Spongebob-patrick-crying.jpg/revision/latest?cb=20140713205315]

---
class: center, middle

# So you work together...

.center[
  ![pair programming](/images/pack-your-bags/working-together.jpg)
]

---
class: center, middle

# Then you help a few more people...

.center[
  ![go team](/images/pack-your-bags/go-team.jpg)
]

---
class: center, middle

# Act Three!

---
class: center, middle

# Suddenly McGuffin has FANS!

.center[
  ![all the github stars!!!](/images/pack-your-bags/your-fans.jpg)
]

---
class: center, middle

# This won't scale...

.center[
  ![nobody wants to do this](/images/pack-your-bags/scroll-of-truth.png)
]

---
class: center, middle

# So what do we do...

.center[
  ![this is my thinking face](/images/pack-your-bags/thinking.jpg)
]

---
class: center

# Containers helped us ship our app...

.center[
  ![ship it](/images/pack-your-bags/container-ship.jpg)
]

---
class: center

# But containers don't really solve this...

.center[
  ![half way there](/images/pack-your-bags/scroll-of-sad-truth.png)
]

---
class: middle

# This is the problem CNAB wants to solve

---

# Hashtag Goals

--

* Package All The Logic To Make Your App Happen

--

* Allow Consumer To Verify Everything They Will Install

--

* Distribute Them In Verifiable Way

---
class: center, middle

# How that works

.center[
  ![workflow](/images/pack-your-bags/the-workflow.png) ![magic](/images/pack-your-bags/magic.gif)
]

.footnote[_http://www.reactiongifs.com/magic-3_]

---
## Try it out: Install a bundle

```console
$ porter install --reference deislabs/porter-hello-velocity:latest
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

--

* MSI for the Cloud
--

* It's a Docker Image
--

* It contains all the tools you need to install your app
--

* It contains configuration, metadata, templates, etc

---
class: center, middle

# The Invocation Image

.center[
  ![so what is it](/images/pack-your-bags/easy-bake-oven-image.png)
]

---

# The Bundle Descriptor
--

* JSON!
--

* List of the invocation image(s) (with digests!)
--

* List of the application image(s) (with digests!)
--

* Definitions for inputs and outputs
--

* Can be signed

---

# Are we done?
--

* We can install (complicated) things

--

* We can verify what we are going to install

--

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

# CNAB Specification
--

* The Bundle format
--

* Defines how things are passed into and out of the invocation image
--

* A required entrypoint in invocation image
--

* Well-defined verbs
--

  * Install
  * Upgrade
  * Uninstall

---
class: center, middle

# An Example: Azure MySQL + Wordpress

.center[
  https://getporter.org/examples/src/azure-wordpress
]

---
name: talkback

# Talk Back

* What tools do you use to deploy?
* Is it to a cloud? On-premise?
* Are you using a mix of tooling and platforms?
* What does your deployment lifecycle look like?

---
class: center, middle

# Manage Bundles with Porter

.center[
  🚨 Not Setup Yet? 🚨

  https://getporter.org/pack-your-bags/#setup
  
  ]
---
name: hello
class: center

# Tutorial
# Hello World

.center[
  ![whale saying hello](/images/whale-hello.png)
]
---

## porter create

```console
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
invocationImage: porter-hello:latest

install:
  - exec:
      description: "Install Hello World"
      command: bash
      flags:
        c: echo Hello World
```

---

## Try it out: porter create

```console
$ mkdir hello
$ cd hello
$ porter create
creating porter configuration in the current directory
$ ls
Dockerfile.tmpl  README.md  porter.yaml
```

---

## porter build

```console
$ porter build --help
Builds the bundle in the current directory by generating a Dockerfile 
and a CNAB bundle.json, and then building the invocation image.
```

---

## Try it out: porter build

```console
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

COPY . /cnab/app
RUN mv /cnab/app/cnab/app/* /cnab/app && rm -r /cnab/app/cnab
# exec mixin has no buildtime dependencies
```
.footnote[🚨 Generated by Porter]

---

### .cnab/
```console
$ tree .cnab/
.cnab
├── app
│   ├── mixins
│   │   └── exec
│   │       ├── exec
│   │       └── exec-runtime
│   ├── porter-runtime
│   └── run
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
    "description": "An example Porter configuration",
    "invocationImages": [{
            "image": "porter-hello:latest",
            "imageType": "docker"
        }],
    "name": "HELLO",
    "definitions": {
        "porter-debug": {
          "type": "boolean"
        }
    },
    "parameters": {
        "fields": {
            "porter-debug": {
                "description": "Print debug information from Porter when executing the bundle",
                "definition": "porter-debug",
                "destination": {
                    "env": "PORTER_DEBUG"
                }
            }
        }
    },
    "version": "0.1.0"
}
```
.footnote[🚨 Generated by Porter]

---

## porter install

```console
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

```console
$ porter install

installing HELLO...
executing porter install configuration from /cnab/app/porter.yaml
Install Hello World
Hello World
execution completed successfully!
```

---
class: center, middle

# BREAK

---
name: mellamo
class: center

# Tutorial
# Hi, My Name is _

.center[
  ![como se llamo me llamo llama](/images/me-llamo.jpg)
]

---
name: parameters

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

```console
$ porter list
NAME          CREATED         MODIFIED        LAST ACTION   LAST STATUS
HELLO_LLAMA   5 seconds ago   3 seconds ago   install       success
HELLO         8 minutes ago   8 minutes ago   install       success
```

???
Ask them to list their bundles

---
name: claims

### Claims

Claims are records of any actions performed by CNAB compliant tools on a bundle.

```console
$ porter show HELLO
  Name: HELLO
  Created: 2019-11-08
  Modified: 2019-11-08
  Last Action: install
  Last Status: success
```

---
name: cleanup-hello
## Cleanup Hello World

First run `porter uninstall` without any arguments:
```console
$ porter uninstall
uninstalling HELLO...
executing porter uninstall configuration from /cnab/app/porter.yaml
Uninstall Hello World
Goodbye World
execution completed successfully!
```

Now run `porter uninstall` with the name you used for the modified bundle:
```console
$ porter uninstall HELLO_LLAMA
uninstalling HELLO_LLAMA...
executing porter uninstall configuration from /cnab/app/porter.yaml
Uninstall Hello llama
Goodbye llama
execution completed successfully!
```

---
name: wordpress
class: center

# Tutorial
# Wordpress

---
name: credentials

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

In all honesty this area is a work in progress. I would shove as everything in a 
credential for now but be aware of the distinction and where the CNAB spec is moving.

---
## porter credentials generate

```console
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
## Try it out: porter credentials generate

Generate a set of credentials for the wordpress bundle in this repository.

1. Change to the `wordpress` directory under the workshop materials
1. Run `porter build`
1. Run `porter credentials generate` and follow the interactive prompts to create a set of credentials
for the wordpress bundle.

???
we all do this together

---
## Try it out: porter install --credential-set

Install the wordpress bundle and pass it the named set of credentials that you generated.

```console
$ porter install --credential-set wordpress
```

---
name: cleanup-wordpress

## Cleanup Wordpress

```console
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
name: "azure-wordpress"
version: "0.1.0"
tag: "deislabs/azure-wordpress"
```

---
## Mixins

Declare any mixins that you are going to use

```yaml
mixins:
  - azure
  - helm3
```

---
## Parameters

Define parameters that the bundle requires.

```yaml
parameters:
- name: mysql_user
  type: string
  default: wordpress
- name: mysql_password
  type: string
  sensitive: true
```

---
## Credentials

Define credentials that the bundle requires and where they should be placed
in the bundle when it is executing:

* environment variables (env)
* files (path)

```yaml
credentials:
- name: SUBSCRIPTION_ID
  env: AZURE_SUBSCRIPTION_ID
- name: kubeconfig
  path: /home/nonroot/.kube/config
```

---
name: dockerfile
## Custom Dockerfile

Specify a custom Dockerfile for the invocation image

* Use a different base image
* Add users, tweak the environment and configuration
* Install tools and applications

⚠️ You are responsible for copying files into the bundle under /cnab/app/

---

### porter.yaml

```yaml
dockerfile: Dockerfile.tmpl
```

### Dockerfile.tmpl

```Dockerfile
FROM debian:stretch-slim

RUN apt-get install -y curl

COPY myscript.sh /cnab/app/
```

---
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
## Review: Default Dockerfile

* Uses Debian for the base image
* Installs root ssl certificates

🎩✨ Automatically copies everything into the bundle under /cnab/app/ for you

---
# Actions and Steps

---
## Actions

Actions map to the verbs you use when you use Porter.

* porter install
* porter upgrade
* porter uninstall

These are defined in the CNAB specification.

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

```console
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
* terraform

.center[ https://getporter.org/mixins ]

---
name: helm
background-image: url(/images/pack-your-bags/helm-mixin.png)

---
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
# Make Your Own Mixin

---
## What makes a mixin?

* Executable written in any language
* Communicates to Porter on stdin and stdout
* Supports a few commands: build, schema, install, upgrade, uninstall
* Translates the steps from the porter manifest to commands against any external tool or service

---
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

# Publishing Your Bundles

Use  `porter publish` to share bundles:

* Publishes both the invocation image and the CNAB bundle manifest to an OCI registry.
* Uses Docker tags for both

```yaml
name: porter-azure-wordpress
version: 0.1.0
invocationImage: deislabs/porter-azure-wordpress:latest
tag: deislabs/porter-azure-wordpress-bundle:latest
```

---

## porter publish

```console
$ porter publish --help
Publishes a bundle by pushing the invocation image and bundle to a registry.


Examples:
  porter publish

```

---

# Try to publish

* Modify the helloworld bundle again
* Edit the porter.yaml:
  * Change the `invocationImage` property to reflect your Docker Hub account
  * Change `tag` property to reflect your Docker Hub account
  * Change the message to something unique
* Run `porter build`
* Run `porter publish`

Example (assuming your username is cnabaholic):

* Change `deislabs/cool-image:latest` to `cnabaholic/cool-image:latest`

---

# Now run your bundle

* Run `porter install --reference [your new tag]`

Example tag of `cnabaholic/hello-people:latest`:

* Run `porter install --reference cnabaholic/hello-people:latest`

---

# How does it do that?

* Updates the bundle with the new images and digests
* Stores the bundle information as a combination:
  * OCI manifest list annotations to the main manifest list for the tag
  * A new manifest list for the bundle credentials and parameters.
* When pulling a bundle, it reconstructs it from the parts mentioned above

See [OCI Bundle Format](/oci-bundle-format) for an example.

---
class: center, middle

# CNAB Ecosystem and Beyond

???
Explain where porter shines, what it is good at vs. say docker app

---

# Next Steps

???
What should someone do if they are interested in CNAB for their work or personal projects?
What is the timeline for the project and how should they be thinking about beginning to incorporate it?

---
class: center, middle

# Choose your own adventure!

* ASCII Art Gophers
* Use Porter with Your Favorite Cloud Provider

---
name: asciiart
# Try it out: ASCII Art Gophers

Make a bundle for the https://github.com/stdupp/goasciiart tool. Use it to
convert cute pictures of gophers into ASCII art when you install the bundle.

Here are some hints so that you can try to solve it in your own way. 
For the full solution, see the [asciiart][asciiart] directory in the workshop materials.

* A good base image for go is `golang:1.11-stretch`.
* You need to run `porter build` after modifying the Dockerfile.tmpl to rebuild
your invocation image to pick up your changes.
* Don't forget to copy your images into your invocation image to /cnab/app/.
* The command to run is `goasciiart -p=gopher.png -w=100`.

[asciiart]: /src/workshop/asciiart

---
name: break-glass
# Use Porter with Your Favorite Cloud Provider

Use the a custom dockerfile template and the exec mixin
to make Porter do something with your favorite cloud provider
such as AWS or GCE.

---
name: rate
class: center, middle
# Workshop Feedback

Please take a minute now to rate this workshop before you leave

<img src="/images/pack-your-bags/feedback-desktop.png" class="left" width="500px" />
<img src="/images/pack-your-bags/feedback-mobile.png" class="right" width="350px" />
