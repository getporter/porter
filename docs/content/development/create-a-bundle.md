---
title: Create a Bundle
description: Create a bundle with Porter
---

Let's walk through how to create and customize your very own Porter bundle.
A bundle includes both the tools and the scripts or logic necessary to automate the deployment.
When writing a bundle, it is best if you have already figured out the commands necessary to perform the deployment first, and then use Porter to package that into a bundle.
Learning Porter while also figuring out how to deploy a particular application can be difficult.

## Requirements
You must [install Porter], and optionally can use the [Porter Visual Studio Code] extension for autocomplete while editing the porter.yaml file.

* [Create a Bundle](#create-a-bundle)
* [Verify the Bundle](#verify-the-bundle)
* [Install Tools](#install-tools)
    * [Use Mixins](#use-mixins)
    * [Use a Custom Dockerfile](#use-a-custom-dockerfile)
* [Customize the Install Action](#customize-the-install-action)
* [Test Your Bundle](#test-your-bundle)
* [Publish Your Bundle](#publish-your-bundle)
* [Use the Published Bundle](#use-the-published-bundle)

## Create a Bundle

Use the [porter create](/cli/porter_create) command to scaffold a new bundle in the current directory.
The directory containing the files for the bundle is called the **bundle directory**.
The generated bundle is very similar to the [hello example bundle] and prints out "Hello World" when installed.
It does not allocate any resources and is safe to run and uninstall when you are finished.

## Verify the Bundle

Your bundle is ready to build and run!
Let's do a quick check before making any further changes to verify that everything is working.

1. Use the [porter build] command to build the bundle. 
   This prepares the bundle so that it can be distributed over a registry by building the bundle image and packaging it in a bundle.
2. Use the [porter install] command to run the bundle's install action defined in the porter.yaml file.

   ```console
   $ porter install mybundle
   executing install action from porter-hello (installation: /mybundle)
   Install Hello World
   Hello World
   execution completed successfully!
   ```

   You do not need to specify the \--reference flag with the install command because the current directory is a bundle directory, containing a porter.yaml file.
   Porter commands that have a \--reference flag use the bundle definition in the current directory when \--reference is not provided.

## Install Tools

Now that you have a working bundle, the next step is to figure out what tools you need installed in the bundle.
Consider what command-line tools that you use today to automate your deployment, such as Terraform, a cloud provider CLI, ansible, etc.
There are two ways to install them into the bundle image. You can either [use mixins](#use-mixins) or install them with by defining a [custom Dockerfile](#use-a-custom-dockerfile) for your bundle:

### Use Mixins

[Mixins] are adapters that makes it easier to work with existing tools within a bundle.
You can use the [porter mixins search] command to find existing mixins to use in your bundle.
A mixin handles installing any required tools for you, and provides an optimized experience for working with that tool in a bundle.
If you are working with the same tool often, eventually you will want to [create a custom mixin] so that you can write installation logic, error handling, common commands a single time and reuse them across your bundles.

To use a mixin in your bundle:

1. Install the mixin on your computer.
   Follow the mixin's instructions to install it using the [porter mixin install] command.
   
   Mixins published by the Porter project can be installed with `porter mixin install NAME`.
   If the mixin is published by a third party, you will need to specify the \--url or \--feed-url flags so Porter knows where to find the mixin.
3. Add the name of the mixin to the **mixins** section in porter.yaml.
   For example, to install the [terraform mixin], you would add an array entry with the value "terraform":
   ```yaml
   mixins:
   - exec
   - terraform
   ```

### Use a Custom Dockerfile

When you are just getting started, it may be easier to start small and use a [custom Dockerfile] instead of creating your own mixins.
By default, Porter generates a Dockerfile for your bundle image automatically.
To use a custom Dockerfile, uncomment the **dockerfile** field in your porter.yaml file to use the template.Dockerfile in your bundle directory instead.

```yaml
dockerfile: template.Dockerfile
```

From there, anything you can do with Docker and [Buildkit], you can do in your custom Dockerfile.
You can install tools, certificates, define environment variables, mount secrets, define build arguments, clone repositories, copy files from your local filesystem, and more.

## Customize the Install Action

The default install action uses the exec mixin to call the helpers.sh script in the bundle directory.
The script prints out "Hello world".

Add an array entry, called a **step**, to the **install** section in porter.yaml and use a mixin to run a command when the bundle is installed.
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

If you are not using custom mixins, put your bundle's logic in an executable, such as bash script or compiled binary, and place it in the bundle directory (next to your porter.yaml file).
During porter build, these files are copied into the bundle's image and can be used when the bundle is run.
You can always use the exec mixin when there isn't an existing mixin, or the mixin doesn't support a particular command that you require.

🙏🏼 Please, [do not embed bash commands] directly in the porter.yaml file because it is much more difficult to get the escaping and quotes correct than putting the bash in a separate file.

Bundles actions should be idempotent, meaning an action can be repeated without causing errors, and it should have the same effect each time.
A user of your bundle should be able to re-run install/upgrade/uninstall multiple times in a row, perhaps because the bundle failed half-way through when run the first time.
Make sure that your commands gracefully handle resources already existing in the install and upgrade actions, and handles already deleted resources in the uninstall action.
Some mixins, like the exec mixin, have [built-in support for error handling][ignore-errors] and ignoring errors in certain circumstances.

## Test Your Bundle

After you have finished editing the porter.yaml, repeat the `porter build` command to re-build the bundle with your latest changes.
Then run `porter install mybundle --force` to install the bundle.
The \--force flag is only safe to use in development, and it allows you to incrementally develop a bundle and re-install it without having to first uninstall it and start over after every change.

## Publish Your Bundle

When you are ready to share your bundle with others, the next step is to publish it to a registry.
Most registries work with Porter, if you run into trouble check our list of [compatible registries]. 

The [porter publish] command by default pushes the bundle to the registry defined in the porter.yaml file.

1. Edit the registry field and change it to a registry that you can push to.
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

## Use the Published Bundle

Once your bundle is published, people can use it by setting the \--reference flag on relevant porter commands to the bundle's reference.
The name, registry, and version fields are used to generate the bundle's default publish location when porter publish is run.
By default, the bundle is published to REGISTRY/BUNDLE_NAME:vBUNDLE_VERSION.
The destination may be changed by specifying  \--registry, \--reference, or \--tag during [publish](/cli/porter_publish/).
The publish command prints out the full bundle reference when it completes.

For example, the following porter.yaml file would result in the bundle being published to ghcr.io/getporter/porter-hello:v0.3.0.
Note that even if you did not specify the bundle version with a v prefix, in the example below the version is `0.3.0`, by default Porter will use a v prefix in the tag of the bundle reference.

```yaml
name: porter-hello
registry: ghcr.io/getporter
version: 0.3.0
```

Once you have figured out the reference to your published bundle, the best way to verify that it was published successfully is with the [porter explain] command:

```console
# porter explain REFERENCE
$ porter explain ghcr.io/getporter/porter-hello:v0.2.0
```

## Next Steps

Now that you know how to create a bundle, here are some more detailed topics on how to customize and distribute it:

* [Control how your bundle's image is built with a custom Dockerfile](/bundle/custom-dockerfile/)
* [Customize your Porter manifest, porter.yaml][manifest]
* [Porter Manifest File Format](/bundle/manifest/file-format/)
* [Best Practices for the exec Mixin](/best-practices/exec-mixin/)
* [Understand how bundles are distributed](/distribute-bundles/)

[install Porter]: /install/
[Porter Visual Studio Code]: https://marketplace.visualstudio.com/items?itemName=getporter.porter-vscode
[hello example bundle]: /examples/hello/
[manifest]: /bundle/manifest/
[local-registry]: https://docs.docker.com/registry/deploying/#run-a-local-registry
[porter create]: /cli/porter_create/
[porter build]: /cli/porter_build/
[porter publish]: /cli/porter_publish/
[porter install]: /cli/porter_install/
[porter mixins search]: /cli/porter_mixins_search/
[porter explain]: /cli/porter_explain/
[porter mixin install]: /cli/porter_mixin_install/
[Mixins]: /mixins/
[create a custom mixin]: /mixin-dev-guide/
[terraform mixin]: /mixins/terraform/
[do not embed bash commands]: /best-practices/exec-mixin/
[ignore-errors]: /blog/ignoring-errors/
[compatible registries]: /compatible-registries/
[custom Dockerfile]: /bundle/custom-dockerfile/
[Buildkit]: https://docs.docker.com/develop/develop-images/build_enhancements/
