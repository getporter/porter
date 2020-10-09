---
title: QuickStart Guide
descriptions: Get started using Porter
---

So you're interested in learning more about Porter? Great! This guide will walk you through key concepts of managing bundles. You will use the porter CLI to install, upgrade, and uninstall the bundle. 

## Pre-requisites

Docker is currently a prerequisite for using Porter. Docker is used to package up the bundle. 

If you do not have Docker installed, go ahead and [get Docker](https://docs.docker.com/get-docker/). 

## Getting Porter

Next, you need Porter. Follow the Porter [installation instructions](/install/).

## Porter Key Concepts 

For this quickstart, the main concepts that you will use include:

* Bundle - A bundle is your application artifact, client tools, configuration and deployment logic packaged together. 
* Installation - An instance of a bundle installed to your system.
* Tag - The metadata that contains information about the registry, image name, and version. 
* Registry - An OCI compliant artifact store that allows machine image management. 

## Inspect a bundle

Before using a bundle that you've found, you can inspect a bundle to see more information about it with `porter inspect`. By inspecting the bundle you can see all images that will be used as a result of bundle actions. 

```
porter inspect --tag getporter/porter-hello:v0.1.0
```

Sample output:
```
Name: HELLO
Description: An example Porter configuration
Version: 0.1.0

Invocation Images:
Image                                                                                            Type     Digest                                                                    Original Image
getporter/porter-hello@sha256:678f5726e2d79263a4ac3f02a35eaf41a312aa833c6f55afa45155730db375a7   docker   sha256:678f5726e2d79263a4ac3f02a35eaf41a312aa833c6f55afa45155730db375a7   getporter/porter-hello-installer:0.1.0

Images:
No images defined
```

With this information, you can pull, inspect, and vet images before you use the bundle. In this example, you are inspecting the HELLO bundle with the tag  getporter/porter-hello:v0.1.0. There are no referenced images. 

## Install the bundle from a registry

To install a bundle, you use the `porter install` command. 

```
porter install --tag getporter/porter-hello:v0.1.0

```

In this example, you are installing athe 0.1.0 version of the bundle from the default registry by specifying a tag. 

## List 

To see all of the bundles installed, you can use the `porter list` command. 

```
porter list
```

Sample Output:
```
NAME       CREATED          MODIFIED         LAST ACTION   LAST STATUS
HELLO      21 minutes ago   21 minutes ago   install       succeeded
```

In this example, it shows the bundle metadata along with the creation time, modification time, the last action that was performed, and the status of the last action. The default name of the bundle, HELLO, is used since there was no installation name specified at the command line. 

## Show

To see information about a specific bundle after it's installed, use the `porter show` command with the name of the bundle.

```
porter show HELLO
```

Sample Output:
```
Name: HELLO
Created: 48 minutes ago
Modified: 36 minutes ago

History:
----------------------------------------
  Action     Timestamp       Status
----------------------------------------
  install    48 minutes ago  succeeded
  uninstall  42 minutes ago  succeeded
  install    36 minutes ago  succeeded
  ```


## Upgrade the bundle

If a bundle is updated and you want to install a later version, you can update the bundle using `porter upgrade`.

```
porter upgrade --tag getporter/porter-hello:v0.1.1
```

Sample Output:
```
upgrading HELLO...
executing upgrade action from HELLO (installation: HELLO)
Upgrade Hello World
Upgraded to World 2.0
execution completed successfully!
```

## Cleanup

To uninstall a bundle, use the `porter uninstall` command. 

```
porter uninstall HELLO
```

## Next Steps 

So in this quickstart, you learned how to use some of the features of the porter cli to manage and inspect bundles. Next, learn more about using Porter with parameters.
