---
title: Manage Installations
description: Manage bundle installations with Porter
---

When a bundle is installed, an **installation** is created.
Porter tracks this installation in its database.
You can use [porter installation list] to see a list of installations and [porter installation show] to view details of a particular installation.

Installations may be scoped in a namespace, which allows you to group related installations together.
Installation names must be unique within a namespace.
Installations that are not defined in a namespace are considered global, and may be referenced both by other global resources and namespaced resources.

There are two ways to manage installations: [imperative commands](#imperative-commands) or [desired state](#desired-state). 
They are not mutually exclusive, and you can switch back and forth between them at any time.

## Imperative Commands

**Imperative** commands such as [porter install] or [porter upgrade].
This is a great way to try out Porter and become familiar with a new bundle when testing it locally.
You can also use it to automate Porter, if that fits into your workflows better than desired state.
In this mode, Porter handles creating and modifying the installation resources for you.

You can use the following command to export the current installation definition to a file.
This is one way to create or update an installation file without figuring out what should go in the file.

```
porter installation show NAME --output yaml 1>installation.yaml
```

## Desired State

**Desired State** commands, such as [porter installation apply], where you are responsible for specifying the _desired state_ of the installation within a file,
and Porter determines the appropriate action to take (if any) to synchronize the installation status to that state.
For example, applying an installation file for a bundle that has not yet been installed will result in Porter executing the install action.
Re-applying that same file would result in Porter doing nothing, because the bundle is already in its desired state.
However, if you changed a parameter value for the installation, or the bundle version, then Porter would execute the upgrade action.

The following will result in Porter executing the bundle:
* The installation has not completed successfully yet.
* The bundle reference has changed. The bundle reference is resolved using the bundleRepository, bundleVersion, bundleDigest, and bundleTag fields.
* The resolved parameter values have changed, either because an associated parameter set has changed, the parameters defined on the bundle have changed, or the values resolved by any parameter sets have changed.
* The list of credential set names have changed. Currently, Porter does not compare resolved credential values.
* The porter installation apply command was run with the --force.

Allowing Porter to manage reconciling the state of the installation is how the [Porter Operator] will work when it is ready, and is well suited for use with GitOps.
With a GitOps workflow, you define the desired state of your applications and infrastructure in code, check it into version control (git), and then trigger workflows when those files are modified. 

## Next Steps

* [Install a bundle using imperative commands with the](/quickstart/)
* [Define an installation in a file and manage it using desired state](/quickstart/desired-state/)
* [Reference: File Formats](/reference/file-formats/)

[porter installation list]: /cli/porter_installations_list/
[porter installation show]: /cli/porter_installations_show/
[porter install]: /cli/porter_install/
[porter upgrade]: /cli/porter_upgrade/
[porter installation apply]: /cli/porter_installations_apply/
[Porter Operator]: /operator/
