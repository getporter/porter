---
title: "Docker Mixin for Porter!"
description: New mixin to use the Docker CLI within the Porter manifest
date: "2020-07-28"
authorname:  Gauri Madhok
author: "@gaurimadhok"
authorlink: "https://twitter.com/gaurimadhok"
authorimage: "https://github.com/gaurimadhok.png"
tags: ["docker", "mixins"]
---

We are excited to announce the first release of a Docker mixin for Porter! 🐳

Mixins are critical building blocks for bundles, and we hope the Docker mixin will help ease the process of composing bundles. The Docker mixin installs Docker and provides the Docker CLI within bundles. Prior to the creation of this mixin, in order to use Docker within your bundle, you would have to create a custom Dockerfile and install Docker. Then, to run any Docker commands, you would need to use the exec mixin and call a bash script to execute the Docker commands. 

The Docker mixin abstracts this logic for you and allows you to specify the Docker commands with the arguments and flags that you want to execute directly in the Porter manifest. The commands currently provided by the Docker mixin are pull, push, run, build, login, and remove. To view all the syntax for the commands, take a look at the [README](https://github.com/deislabs/porter-docker).

Let's go through an example bundle to try out the mixin. First, we will use the Docker mixin to pull and run [docker/whalesay](https://hub.docker.com/r/docker/whalesay/). Then, we will write our own Dockerfile, build it, and push it to Docker Hub.

## Author the bundle
Writing a bundle with the Docker mixin has a few steps:

* [Create a bundle](#create-a-bundle)
* [Install the Docker mixin](#install-the-docker-mixin)
* [Add the Docker mixin to the Porter manifest](#add-the-docker-mixin-to-the-porter-manifest)
* [Use Docker CLI](#use-docker-cli)
* [Set up credentials](#set-up-credentials)

Let's run through these steps with our example bundle called docker-mixin-practice. First, set up a project:
```
mkdir docker-mixin-practice;
cd docker-mixin-practice;
```

### Create a bundle
Next, use the porter create command to generate a skeleton bundle that you can modify as we go through our example. 
```console
$ porter create
```

### Install the Docker mixin
Next, you need to install the Docker mixin to extend the Porter client. To install the mixin, run the line below and you should see the output that it was installed.
```console
$ porter mixins install docker

installed docker mixin v0.1.0 (b660770)
```
This installs the docker mixin into porter, by default in ~/.porter/mixins.

### Add the Docker mixin to the Porter manifest
In order to use the Docker mixin within your bundle, you need to add it to the mixin list. In the create a bundle step, porter create added the exec mixin. Replace the exec line with docker. 
```
mixins:
- docker
```

### Use Docker CLI

Next, delete the install, upgrade, and uninstall actions. Now, to run docker/whalesay, copy and paste the code below into the porter.yaml to pull the image and then run it with a command to say "Hello World". 
```
install:
- docker:
    description: "Install Whalesay"
    pull:
      name: docker/whalesay
      tag: latest
- docker:
    description: "Run Whalesay"
    run:
      name: dockermixin
      image: "docker/whalesay:latest"
      command: cowsay
      arguments:
        - "Hello World"

uninstall:
- docker:
    description: "Remove dockermixin container"
    remove:
      container: dockermixin
```
When you are ready to install your bundle, run the command below to install and give access to the Docker Daemon. 

```console
$ porter install demo --allow-docker-host-access
```
This is the output that should be generated after it runs. 
```
Run Whalesay
 _____________ 
< Hello World >
 ------------- 
    \
     \
      \     
                    ##        .            
              ## ## ##       ==            
           ## ## ## ##      ===            
       /""""""""""""""""___/ ===        
  ~~~ {~~ ~~~~ ~~~ ~~~~ ~~ ~ /  ===- ~~~   
       \______ o          __/            
        \    \        __/             
          \____\______/   
```

Now, we will go through an example of how you can incorporate and build your own Docker image and then push it to Docker hub. First, you will need to create a Dockerfile named Dockerfile-cookies next to your porter.yaml and copy paste the code below into the file.

```
FROM debian:stretch-slim

CMD ["echo", "Everyone loves cookies"]
```
To build an image from this Dockerfile, copy and paste the code below and replace the current install action. This will build your image, login to Docker, and push your image to a registry. 

Change YOURNAME to your docker username. 
```
install:
- docker:
    description: "Build image"
    build:
      tag: "YOURNAME/cookies:v1.0"
      file: Dockerfile-cookies
- docker:
    description: "Login to docker"
    login:
- docker:
    description: "Push image"
    push:
      name: YOURNAME/cookies
      tag: v1.0
```

### Set up credentials
This step is needed if you wish to use Docker push to push an image to a registry. In order to push to one of your registries, you need to login to Docker Hub. To set up your credentials to login to Docker Hub, make sure the environment variables DOCKER_USERNAME and DOCKER_PASSWORD are set on your machine. Next, add these lines in your porter.yaml. You may change the name to what you want it to be.
```
credentials:
  - name: DOCKER_USERNAME
    env: DOCKER_USERNAME
  - name: DOCKER_PASSWORD
    env: DOCKER_PASSWORD
``` 
Next, run the following commands and edit the file with where the credentials will come from.
```console
$ porter credentials create docker.json
$ cat docker.json
# modify docker.json with your editor to the content below
{
    "schemaType": "CredentialSet",
    "schemaVersion": "1.0.1",
    "name": "docker",
    "credentials": [
        {
            "name": "DOCKER_USERNAME",
            "source": {
                "env": "DOCKER_USERNAME"
            }
        },
        {
            "name": "DOCKER_PASSWORD",
            "source": {
                "env": "DOCKER_PASSWORD"
            }
        }
    ]
}
$ porter credentials apply docker.json
```
Your credentials are now set up. When you run install or upgrade or uninstall, you need to pass in your credentials using the `-c` or `--credential-set` flag. 

When you are ready to install your bundle, run the command below to identify the credentials and give access to the Docker daemon. 

```console
$ porter install demo -c docker --allow-docker-host-access
```
After it runs, you should see output that the image was built and tagged successfully, the login succeeded, and the push to your repository happened.
```
Build image
Sending build context to Docker daemon  101.2MB
Step 1/2 : FROM debian:stretch-slim
 ---> 614bb74b620e
Step 2/2 : CMD ["echo", "Everyone loves cookies"]
 ---> Using cache
 ---> e89d85933cc1
Successfully built e89d85933cc1
Successfully tagged *******/cookies:v1.0
Login to docker
Login Succeeded
Push image
The push refers to repository [docker.io/*******/cookies]
6885f9305c0a: Preparing
6885f9305c0a: Layer already exists
v1.0: digest: sha256:1fb89bd28f81c3e29ae16a44f077f4709f33ac410581faf23b42d9bd7a6f913b size: 529
execution completed successfully!
``` 

Finally, run the following command to clean everything up:
```console
$ porter uninstall demo -c docker --allow-docker-host-access
```

## Thank you for reading!
Please try out the mixin and let us know if you have any feedback to make it better! You can dig into the code [here](https://github.com/deislabs/porter-docker)  and create an issue [here](https://github.com/deislabs/porter-docker/issues/new).
