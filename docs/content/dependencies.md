---
title: Dependencies
description: Dependency Management with Porter
---

## Define a dependency

In the manifest, list the dependencies in the order that they should be
installed. Only bundles that have been authored by Porter, and have a Porter
manifest, can be dependencies.

```yaml
dependencies:
- name: mysql
```

Currently dependencies are assumed to be located in one of the locations below.
Once the CNAB spec has a mechanism for distributing bundles via registries, this
will be revisited.

* bundles (in the same directory as the Porter manifest)
* PORTER_HOME/bundles

Parameters defined in a dependent bundle can be defaulted from the root bundle.
In the example below, the mysql bundle defined the `database_name` and
`mysql_user` parameters, and the root bundle (Wordpress) decided to default them
to a specific value, so that the user wouldn't be required to select a value for
those parameters when installing Wordpress.

```yaml
dependencies:
- name: mysql
  parameters:
    database_name: wordpress
    mysql_user: wordpress
```

## Consolidating Parameters and Credentials

Porter consolidates the parameters and credentials from the dependencies and
promotes them to the root bundle. This allows the user executing the bundle to
interact with the parameter or credential from a dependency just as they would
from one defined on the root bundle.

When a parameter or credential has the same type and name, then it is
consolidated. When the same name is used, but not the same type then the bundle
build fails.

_Note: Porter doesn't yet have a conflict resolution mechanism in place to
allow an author to force consolidating parameters/credentials, or resolve a
conflict._

## Dependency Graph

Currently Porter does not build a dependency graph and only supports direct
dependencies. Dependencies of dependencies, transitive dependencies, are
ignored. See [Design: Dependency Graph Resolution](https://github.com/deislabs/porter/issues/69) for more information.

## Dependency Ordering

Depending on the action being performed, Porter handles executing the dependencies
steps either before or after the steps in the root bundle.

* Install - The dependency's steps are executed _before_ the steps in the root bundle.
* Upgrade - The dependency's steps are executed _before_ the steps in the root bundle.
* Uninstall - The dependency's steps are executed _after_ the steps in the root bundle.

## Dependencies and the CNAB Spec

The CNAB Spec that Porter implements does not define how dependencies between
bundles works (or that they can even exist). Porter has a trick for getting
around that: Porter identifies dependencies between bundles and then builds a
single composite bundle during the `porter build` command. At runtime there is a
single invocation image, so that any tool can work with bundles created by
porter.