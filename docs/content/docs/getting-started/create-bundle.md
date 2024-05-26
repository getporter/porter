---
title: Create a Bundle
aliases:
- /getting-started/create-bundle.md
- /getting-started/create-a-bundle/
- /bundle/create/
description: Create a bundle with Porter
weight: 3
---

Bundles include the tools and logic necessary to automate a deployment and including operations on the runtime-environment.
When writing a bundle, it's best to have figured out the workflow to perform the deployment first. Then use Porter to package that into a bundle.

## Requirements

## Requirements
* [Install Porter]
* Optional: [Porter Visual Studio Code] extension

## Steps
- [Create a Bundle](#create-a-bundle)
- [Verify the Bundle](#verify-the-bundle)
- [Install the Bundle](#install-the-bundle)
  - [Use Mixins](#use-mixins)
- [Publish Your Bundle](#publish-your-bundle)
- [Next Steps](#next-steps)
   - [Leverage a different registry](#leverage-a-different-registry)
   - [Use a Custom Dockerfile](#use-a-custom-dockerfile)
   - [Customize the Install Action](#customize-the-install-action)
   - [Test Your Bundle](#test-your-bundle)
   - [Third Party Mixins](#third-party-mixins)

## Create a Bundle

   **Run this in a new or empty directory**

   ```console
   $ porter create
   ```
   ```
   creating porter configuration in the current directory
   ```

   The [porter create](/cli/porter_create/) creates the scaffolding for a new bundle in the *current directory*.
   This makes your current directory a bundle directory. The directory containing the files for the bundle is called the **bundle directory**.
   The generated bundle is very similar to the [hello example bundle] and prints out "Hello World" when installed.
   It does not allocate any resources and is safe to run and uninstall when you are finished.

   Check out what files Porter created in this directory:

   ```console 
   $ ls
   ```
   ```
   README.md           helpers.sh          porter.yaml         template.Dockerfile
   ```
## Verify the Bundle

Your bundle is ready to build and run!

Let's see what this bundle will do:
**Note: We've cropped out just the Porter actions part of the bundle,
there is more that is generated and commented out**

```console
$ cat porter.yaml
```

```yaml 
install:
  - exec:
      description: "Install Hello World"
      command: ./helpers.sh
      arguments:
        - install

upgrade:
  - exec:
      description: "World 2.0"
      command: ./helpers.sh
      arguments:
        - upgrade

uninstall:
  - exec:
      description: "Uninstall Hello World"
      command: ./helpers.sh
      arguments:
        - uninstall
```

All actions (install, upgrade, uninstall)  are utilizing the `helpers.sh` script that was generated when `porter create` was ran. In the `helpers.sh` this is running:

```bash
install() {
  echo Hello World
}

upgrade() {
  echo World 2.0
}

uninstall() {
  echo Goodbye World
}

```

## Build the Bundle 

```console 
$ porter build
```

The [porter build] command prepares the bundle so that it can be distributed over a registry by building the bundle image and packaging it in a bundle.

When running the build command, Porter converts the `porter.yaml` into a CNAB `bundle.json`, and creates a new directory called `/.cnab` to store all this information:

```console 
$ ls -lah
```
```
schristoff  staff   160B Mar 14 09:57 .cnab
```



## Install the Bundle
Use the [porter install] command to run the bundle's install action defined in the `porter.yaml` file.

   ```console
   $ porter install mybundle
   ```
   ```
   executing install action from porter-hello (installation: /mybundle)
   Install Hello World
   Hello World
   execution completed successfully!
   ```


### Use Mixins

```console 
$ porter mixins search
```

The [porter mixins search] command shows existing mixins to use in your bundle.
[Mixins] are adapters that makes it easier to work with existing tools within a bundle.

Let's add a mixin to the bundle!

```console 
$ porter mixin install docker
```

Edit the `porter.yaml` to have the `docker` mixin added, to add `- docker` underneath `- exec`:

```yaml
schemaType: Bundle
schemaVersion: 1.0.1
name: porter-hello
version: 0.1.0
description: "An example Porter configuration"
registry: "localhost:5000"

mixins:
  - exec
  - docker

```

Let's add the Docker mixin to the install action:

```yaml
mixins:
  - exec
  - docker

install:
  - exec:
      description: "Install Hello World"
      command: ./helpers.sh
      arguments:
        - install
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
```
Great! Now we can build and run install to see the Docker mixin at work

```console 
$ porter install 
```

```console
$ porter install demo --allow-docker-host-access
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
execution completed successfully!
```


## Publish Your Bundle

We will be starting up a local Docker registry for this guide, but there are many [compatible registries] Porter supports.

**Note for MacOS users: AirPlay receiver is by default on and bound to port 5000 Macs, go to System Settings and Disable "AirPlay Receiver" to not get an error here**

```console 
$ docker run -d -p 5000:5000 --name registry registry:3
```

```console 
$ porter publish
```

```console
Rewriting CNAB bundle.json...
Starting to copy image localhost:5000/porter-hello@sha256:c9e80ad...
Completed image localhost:5000/porter-hello@sha256:c9e80adbb1cf82de62... copy
Bundle localhost:5000/porter-hello:v0.1.0 pushed successfully, with digest "sha256:95a3d2...
```

We will be able to see our bundle in our registry here:
```console
$ docker images
```
```
REPOSITORY                    TAG                                       IMAGE ID  
localhost:5000/porter-hello   porter-37da5464f8517662657529ad34851db9   a01f80
```


The [porter publish] command by default pushes the bundle to the registry defined in the `porter.yaml` file.


The name, registry, and version fields are used to generate the bundle's default publish location when porter publish is run.
By default, the bundle is published to REGISTRY/BUNDLE_NAME:vBUNDLE_VERSION.
In the case used above, this looks like:

```console
$ porter explain localhost:5000/porter-hello:v0.1.0
````

This is generated from these fields in the `porter.yaml`

```yaml
name: porter-hello
registry: "localhost:5000"
version: 0.1.0
```

The destination may be changed by specifying \--registry, \--reference, or \--tag during [publish](/cli/porter_publish/).
The publish command prints out the full bundle reference when it completes.


## Next Steps

Now that you know how to create a bundle you can follow the optional steps below, or check out more detailed topics on how to customize and distribute it:

- [Next: What is a bundle?](/quickstart/bundles/)
- [Next: Work with Mixins](/how-to-guides/work-with-mixins/)

### Leverage a different registry

1. Edit the registry field in the `porter.yaml` and change it to a registry that you can push to.
   For example, if you have an account on Docker Hub, change the registry value from localhost:5000 to your Docker Hub username.
2. Edit the name field and change it to your preferred name for the bundle, like "mybundle".
3. Use the docker login command to first authenticate to the destination registry:

   ```
   docker login REGISTRY
   ```

   For example, if the registry defined in the porter.yaml is ghcr.io/myuser, then run `docker login ghcr.io` to authenticate.

   If you are publishing to Docker Hub, the registry field in the bundle would just be your Docker Hub username, and you would authenticate to the registry with just `docker login` without any additional arguments.
   This works because the Docker client by default uses Docker Hub (docker.io) when a registry is not fully specified.

4. Now, publish the bundle by running `porter publish`.

### Use a Custom Dockerfile

To use a custom Dockerfile, uncomment the **dockerfile** field in your porter.yaml file to use the template.Dockerfile in your bundle directory instead.

```yaml
dockerfile: template.Dockerfile
```

From there, anything you can do with Docker and [Buildkit], you can do in your custom Dockerfile.
You can install tools, certificates, define environment variables, mount secrets, define build arguments, clone repositories, copy files from your local filesystem, and more.

### Customize the Install Action

The default install action uses the exec mixin to call the `helpers.sh` script in the bundle directory.
The script prints out "Hello world".

Add an entry to the **install** section in porter.yaml and use a mixin to run a command when the bundle is installed. This is called a **step**.
The syntax for each mixin is different so reference the documentation for your particular mixin to know what to specify.
The general form for any mixin is as follows:

```yaml
install:
  - exec:
      description: Optional description of the step
      # ... mixin specific values
  - terraform:
      description: Optional description of the step
      # ... mixin specific values
```

Bundles actions should be idempotent, meaning an action can be repeated without causing errors, and it should have the same effect each time.
A user of your bundle should be able to re-run install/upgrade/uninstall multiple times in a row, perhaps because the bundle failed half-way through when run the first time.
Make sure that your commands gracefully handle resources already existing in the install and upgrade actions, and handles already deleted resources in the uninstall action.
Some mixins, like the exec mixin, have [built-in support for error handling][ignore-errors] and ignoring errors in certain circumstances.

🚨 [Do not embed bash commands] directly in the porter.yaml file because it is much more difficult to get the escaping and quotes correct than putting the bash in a separate file. 🚨 

If you are not using custom mixins, put your bundle's logic in an executable, such as bash script or compiled binary, and place it in the bundle directory (next to your porter.yaml file).
During porter build, these files are copied into the bundle's image and can be used when the bundle is run.
You can always use the exec mixin when there isn't an existing mixin, or the mixin doesn't support a particular command that you require.

### Test Your Bundle

After you have finished editing the porter.yaml, repeat the `porter build` command to re-build the bundle with your latest changes.
Then run `porter install mybundle --force` to install the bundle.
The \--force flag is only safe to use in development, and it allows you to incrementally develop a bundle and re-install it without having to first uninstall it and start over after every change.


### Third Party Mixins 

   Mixins published by the Porter project can be installed with `porter mixin install NAME`.
   If the mixin is published by a third party, you will need to specify the \--url or \--feed-url flags so Porter knows where to find the mixin.
   A mixin published by Porter would look like:
      ```console
      porter mixin install kubernetes --version canary --url https://cdn.porter.sh/mixins/kubernetes
      ```
   A third party mixin install would look like:
      ```console
      porter mixin install helm3 --feed-url https://mchorfa.github.io/porter-helm3/atom.xml
      ```

[install Porter]: /install/
[Porter Visual Studio Code]: https://marketplace.visualstudio.com/items?itemName=getporter.porter-vscode
[hello example bundle]: /docs/references/examples/hello/
[manifest]: /bundle/manifest/
[local-registry]: https://docs.docker.com/registry/deploying/#run-a-local-registry
[porter create]: /cli/porter_create/
[porter build]: /cli/porter_build/
[porter publish]: /cli/porter_publish/
[porter install]: /cli/porter_install/
[porter mixins search]: /cli/porter_mixins_search/
[porter explain]: /cli/porter_explain/
[porter mixin install]: /cli/porter_mixins_install/
[Mixins]: /mixins/
[create a custom mixin]: /mixin-dev-guide/
[terraform mixin]: /mixins/terraform/
[do not embed bash commands]: /docs/best-practices/exec-mixin/
[ignore-errors]: /blog/ignoring-errors/
[compatible registries]: /compatible-registries/
[custom Dockerfile]: /bundle/custom-dockerfile/
[Buildkit]: https://docs.docker.com/develop/develop-images/build_enhancements/
