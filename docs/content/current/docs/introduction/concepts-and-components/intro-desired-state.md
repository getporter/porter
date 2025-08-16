---
title: Desired State
description: TODO
weight: 5
---

**Desired State** commands, such as [porter installation apply], where you are responsible for specifying the _desired state_ of the installation within a file,
and Porter determines the appropriate action to take (if any) to synchronize the installation status to that state.
For example, applying an installation file for a bundle that has not yet been installed will result in Porter executing the install action.
Re-applying that same file would result in Porter doing nothing, because the bundle is already in its desired state.
However, if you changed a parameter value for the installation, or the bundle version, then Porter would execute the upgrade action.

The following will result in Porter executing the bundle:

- The installation has not completed successfully yet.
- The bundle reference has changed. The bundle reference is resolved using the repository, version, digest, and tag fields.
- The resolved parameter values have changed, either because an associated parameter set has changed, the parameters defined on the bundle have changed, or the values resolved by any parameter sets have changed.
- The list of credential set names have changed. Currently, Porter does not compare resolved credential values.
- The porter installation apply command was run with the --force.

Allowing Porter to manage reconciling the state of the installation is how the [Porter Operator] will work when it is ready, and is well suited for use with GitOps.
With a GitOps workflow, you define the desired state of your applications and infrastructure in code, check it into version control (git), and then trigger workflows when those files are modified.

## Next Steps

- [Install a bundle using imperative commands with the Porter CLI](/quickstart/)
- [Define an installation in a file and manage it using desired state](/quickstart/desired-state/)
- [Reference: File Formats](/reference/file-formats/)
