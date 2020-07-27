---
title: Build your Bundle
description: Building your first bundle
---

Once you have a [porter.yaml](porter-yaml) manifest, it's time to build your first bundle. 

The `porter build` command generates this bundle:

```
$ porter build
Copying porter runtime ===>
Copying mixins ===>
Copying mixin exec ===>

Generating Dockerfile =======>

Writing Dockerfile =======>

Starting Invocation Image Build =======>
```

Let's look at these steps in more detail. 

### First, Porter will copy its runtime plus any mixins into the `.cnab/app` directory of your bundle. If you look in the .cnab directory, you'll see something like this 

```.
├── app
│   ├── mixins
│   │   └── exec
│   │       ├── exec
│   │       └── exec-runtime
│   ├── porter-runtime
│   └── run
└── bundle.json
```

Remember the manifest includes the lines:

```
mixins:
  - exec
  ```

Porter locates the available mixins in the `$PORTER_HOME/mixins` directory. By default, the Porter home directory is located in `~/.porter`. In this example, we are using the `exec` mixin, so the `$PORTER_HOME/mixins/exec` directory will be copied. 

### After copying any mixins to the `.cnab` directory of the bundle, a Dockerfile is generated. 

Dockerfiles are text files that contain all the commands needed to build a given image, or in this case define what is needed to package up the bundle. 

Back to the manifest and these lines:

```
# Uncomment the line below to use a template Dockerfile for your invocation image
#dockerfile: Dockerfile.tmpl
```

Since we didn't configure a Dockerfile.tmpl and update the manifest a default base image is used:

```
FROM debian:stretch

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates
```

This set of instructions:

* starts from the [debian:stretch Docker image](https://hub.docker.com/_/debian). 
* defines that a BUNDLE_DIR variable will be passed in at build-time
* executes "apt-get update" which downloads package information from all configured sources
* executes "apt-get install -y ca-certificates" if the previous update is successful which installs the latest CA certificates. 

Porter will check any included mixins for build dependencies and add these as well. For the exec mixin there are no buildtime dependencies. 

Then the rest of the default base image is used:

```

COPY . $BUNDLE_DIR
RUN rm -fr $BUNDLE_DIR/.cnab
COPY .cnab /cnab
COPY porter.yaml $BUNDLE_DIR/porter.yaml
WORKDIR $BUNDLE_DIR
CMD ["/cnab/app/run"]
```

This set of instructions:

* copies from current directory . into the BUNDLE_DIR
* executes "rm -fr $BUNDLE_DIR/.cnab" which removes the previous .cnab directory from the bundle
* copies from the current directory the .cnab directory into the BUNDLE_DIR
* copies the manifest into $BUNDLE_DIR/porter.yaml
* sets up the working directory to $BUNDLE_DIR for all subsequent commands. 
* sets up "/cnab/app/run" as the default command.

### Once the Dockerfile is generated, it's written. 

### Finally, the package is built. 

```
Step 1/9 : FROM debian:stretch
 ---> 5738956efb6b
Step 2/9 : ARG BUNDLE_DIR
 ---> Using cache
 ---> c9d91881dd7c
Step 3/9 : RUN apt-get update && apt-get install -y ca-certificates
 ---> Using cache
 ---> afa85b98ed97
Step 4/9 : COPY . $BUNDLE_DIR
 ---> Using cache
 ---> e4057b41978c
Step 5/9 : RUN rm -fr $BUNDLE_DIR/.cnab
 ---> Using cache
 ---> ee114d95bc2d
Step 6/9 : COPY .cnab /cnab
 ---> Using cache
 ---> 1bb73c63ef65
Step 7/9 : COPY porter.yaml $BUNDLE_DIR/porter.yaml
 ---> Using cache
 ---> 483c6b05a0b7
Step 8/9 : WORKDIR $BUNDLE_DIR
 ---> Using cache
 ---> 9d2497296f3b
Step 9/9 : CMD ["/cnab/app/run"]
 ---> Using cache
 ---> 23c208fd5dc7
Successfully built 23c208fd5dc7
Successfully tagged getporter/porter-hello-installer:0.1.0
DEBUG name:    arm
DEBUG pkgDir: /Users/sigje/.porter/mixins/arm
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/arm/arm version --output json --debug
DEBUG name:    aws
DEBUG pkgDir: /Users/sigje/.porter/mixins/aws
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/aws/aws version --output json --debug
DEBUG name:    az
DEBUG pkgDir: /Users/sigje/.porter/mixins/az
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/az/az version --output json --debug
DEBUG name:    exec
DEBUG pkgDir: /Users/sigje/.porter/mixins/exec
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/exec/exec version --output json --debug
DEBUG name:    gcloud
DEBUG pkgDir: /Users/sigje/.porter/mixins/gcloud
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/gcloud/gcloud version --output json --debug
DEBUG name:    helm
DEBUG pkgDir: /Users/sigje/.porter/mixins/helm
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/helm/helm version --output json --debug
DEBUG name:    kubernetes
DEBUG pkgDir: /Users/sigje/.porter/mixins/kubernetes
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/kubernetes/kubernetes version --output json --debug
DEBUG name:    terraform
DEBUG pkgDir: /Users/sigje/.porter/mixins/terraform
DEBUG file:
DEBUG stdin:

/Users/sigje/.porter/mixins/terraform/terraform version --output json --debug
```

