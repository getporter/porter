---
title: "Migrate from Docker App to Porter"
description: "How to migrate from Docker App to Porter"
date: "2021-05-05"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://twitter.com/carolynvs"
authorimage: "https://github.com/carolynvs.png"
tags: ["docker", "docker-app"]
image: "images/migrate-from-docker-app-twitter-card.png"
summary: | 
  Docker recently announced that they are no longer developing Docker App, and that you should migrate to Porter to continue using your app and work with it like you do today. Let's walk through how to migrate your Docker App to Porter.
---

<img src="/images/porter-with-docker.png" width="250px" align="right"/>

Welcome Docker App users! 🎉
Docker recently [announced] that they are no longer developing Docker App, and that you should migrate to Porter to continue using your app (bundle) and work with it like you do today.
Let's walk through how to migrate your Docker App to Porter.


Docker App under the hood has always created and deployed Cloud Native Application Bundles, though you may not have realized it.
Porter also supports creating bundles, publishing them to OCI registries, and managing the application lifecycle with install, upgrade and uninstall commands.

Porter uses mixins to compose bundles using existing tools, such as docker-compose, helm, or terraform.
We will use the [docker-compose mixin] to migrate an existing Docker App to Porter.

1. [Install Porter].

1. Install the [docker-compose mixin] by running `porter mixins install docker-compose`.

1. Create a porter.yaml file in the same directory as your docker app next to the docker-compose.yaml.
    Copy into porter.yaml the contents below.

    <script src="https://gist-it.appspot.com/https://github.com/getporter/porter/blob/main/examples/dockerapp/porter.yaml"></script>

1. Open metadata.yaml from your docker app.

    ```yaml
    version: 0.1.0
    name: my-docker-app
    description: My amazing docker app
    maintainers:
      - name: Maria Gomez
        email: mariagomez@example.com
    ```

1. Open porter.yaml and copy the version, name, and description values into porter.yaml into the corresponding fields.
    Porter doesn't have a field for maintainers, so that doesn't need to be migrated.

    ```yaml
    name: my-docker-app
    version: 0.1.0
    description: My amazing docker app
    ```

1. Open your parameters.yaml. Add each parameter to porter's parameters field.
    If your parameter used a period or other characters that are not allowed in an environment variable name, replace that character with an acceptable substitute such as underscore _.
    Update your docker-compose.yaml to use any of the newly renamed parameters.

    **parameters.yaml**
    ```yaml
    hello:
      text: hello from porter
      porter: 8080
    ```

    **porter.yaml**
    ```yaml
    parameters:
    - name: hello_text
      type: string
      env: hello_text
      default: hello from porter
    - name: hello_port
      type: integer
      env: hello_port
      default: 8080
    ```

1. Install the bundle with `porter install --allow-docker-host-access`.
   The `--allow-docker-host-access` flag is required so that the bundle can communicate with the docker host.
  
    Porter supports the [DOCKER_HOST and DOCKER_CONTEXT environment variables](https://www.docker.com/blog/how-to-deploy-on-remote-docker-hosts-with-docker-compose/).
    You can use these to have Porter deploy your application to a remote host.

    ```console
    $ porter install --allow-docker-host-access
    installing my-docker-app...
    executing install action from my-docker-app (installation: my-docker-app)
    /usr/local/lib/python3.5/dist-packages/paramiko/transport.py:33: CryptographyDeprecationWarning: Python 3.5 support will be dropped in the next release of cryptography. Please upgrade your Python.
      from cryptography.hazmat.backends import default_backend
    Creating app_hello_1 ... done
    execution completed successfully!
    ```

1. Confirm that your application was deploy with `docker ps`.

    ```console
    $ docker ps
    CONTAINER ID   IMAGE                 COMMAND                  CREATED          STATUS          PORTS                                       NAMES
    c5428e359333   hashicorp/http-echo   "/http-echo -text 'h…"   27 minutes ago   Up 27 minutes   0.0.0.0:8080->5678/tcp, :::8080->5678/tcp   app_hello_1
    ```

1. You can view your installations with `porter list`:
  
    ```console
    $ porter list
    NAME                 CREATED          MODIFIED         LAST ACTION   LAST STATUS
    my-docker-app        28 minutes ago   28 minutes ago   install       succeeded
    ```

1. Let's look at the details of your migrated application with `porter show`.
    The output tells us that it was installed successfully and shows the history of changes made to the installation.

    ```console
    $ porter show my-docker-app
    Name: my-docker-app
    Created: 28 minutes ago
    Modified: 28 minutes ago

    Outputs:
    -------------------------------------------------------------------------------
      Name                                 Type    Value
    -------------------------------------------------------------------------------
      io.cnab.outputs.invocationImageLogs  string  executing install action from
                                                  my-docker-app (installation...

    History:
    ----------------------------------------------------------------------------
      Run ID                      Action   Timestamp       Status     Has Logs
    ----------------------------------------------------------------------------
      01F4YMH7AETP2P38Y81YVQ5TJS  install  28 minutes ago  succeeded  true
    ```

1. So far we have been working inside the "developer iteration loop", where you can edit the bundle on your local filesystem and deploy it to your developer environment to test it.
    Once the bundle is stable, the next step is to publish it to an OCI registry so that others can install your bundle using its reference.
    All of the porter commands accept a flag, \--reference, for example `porter install --reference ghcr.io/getporter/examples/porter-hello:v0.2` so that you do not need to distribute the bundle files themselves.

1. When you are ready to share your bundle with others, select which OCI registry where you will host the bundle, for example, `ghcr.io/getporter` or on Docker Hub under your username `carolynvs`.
    Edit your porter.yaml and set the registry field to the destination registry.
  
    ```yaml
    name: my-docker-app
    version: 0.1.0
    description: My amazing docker app
    registry: carolynvs
    ```
  
1. Publish your bundle to the destination registry with `porter publish`.

    ```console
    $ porter publish
    Pushing CNAB invocation image...
    The push refers to repository [docker.io/carolynvs/my-docker-app-installer]
    a5fd17ef8522: Preparing
    878a51fed4d7: Preparing
    d774d6a15e77: Preparing
    5ff217bf43f5: Preparing
    739733239466: Preparing
    d7b369a46116: Preparing
    356db0d0a1d7: Preparing
    d7b369a46116: Waiting
    356db0d0a1d7: Waiting
    878a51fed4d7: Pushed
    d774d6a15e77: Pushed
    356db0d0a1d7: Mounted from library/debian
    d7b369a46116: Pushed
    a5fd17ef8522: Pushed
    5ff217bf43f5: Pushed
    739733239466: Pushed
    v0.1.0: digest: sha256:b3e56730d60e1f587ba34f4316bc20b22a7f8b1daf13560bf91d67a72b858243 size: 1792

    Rewriting CNAB bundle.json...
    Starting to copy image carolynvs/my-docker-app-installer:v0.1.0...
    Completed image carolynvs/my-docker-app-installer:v0.1.0 copy
    Bundle tag docker.io/carolynvs/my-docker-app:v0.1.0 pushed successfully, with digest "sha256:5631fb62f7fbf3fa7c54fe640808045bd3a7de9cfd691645c48c929398f31e92"
    ```

    The last line of the output prints the full reference to the published bundle, in this case `docker.io/carolynvs/my-docker-app:v0.1.0`.
    You can use or omit the docker.io registry depending on your preference.

1. Now that your bundle is published, let's install it.
    First change your current directory in your terminal to leave the directory containing your bundle's source code.
    We are going to install the bundle again, this time using the published bundle.

    Replace `YOUR_BUNDLE_TAG` with the reference printed by porter publish.
    We will use the installation name `my-app` to reference the installation in subsequent commands.

    ```
    $ porter install my-app --reference YOUR_BUNDLE_REFERENCE
    ```

    For example,
    ```console
    $ porter install my-app --reference carolynvs/my-docker-app:v0.1.0
    ```

1. Oops! You made a mistake in your original bundle and need to fix it.
    Open your porter.yaml file and increase the version number.

    ```
    version: 0.1.1
    ```

1. Rebuild the bundle and publish the new version.

    ```
    porter build
    porter publish
    ```

1. You can use `porter upgrade` to upgrade the installation to the latest version, specifying the newly published version of your bundle.

    ```
    $ porter upgrade my-app --reference YOUR_UPDATED_BUNDLE_REFERENCE --allow-docker-host-access
    ```

    For example:
    ```console
    $ porter upgrade my-app --reference carolynvs/my-docker-app:v0.1.1 --allow-docker-host-access
    ```

    You can configure Porter to always [allow Docker host access] so that you do not need to set it with a flag on every command.

1. Once you are done, uninstall the bundle with `porter uninstall`.
  
    ```
    porter uninstall my-app
    ```

Hopefully the migration process isn't too complicated!
We would love to have you migrate your Docker App to Porter and continue to use CNAB and bundles to manage your applications.
Please [let us know][contact] how the migration went (good or bad), and we are happy to help if you have questions, or you would like help with your migration.

[announced]: https://github.com/docker/roadmap/issues/209
[Install Porter]: /install/
[docker-compose mixin]: /mixins/docker-compose/
[allow Docker host access]: /configuration/#allow-docker-host-access
[contact]: /community/