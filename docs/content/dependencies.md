---
title: Dependencies
description: Dependency Management with Porter
---

In the Porter manifest you can [define a dependency](#define-a-dependency) on another 
Porter authored bundle, and when you build the bundle, the root bundle and its dependencies 
are combined into a single composite bundle. The final bundle is a CNAB compliant bundle
that can be used with any CNAB tool.

Since the CNAB Spec does not allow for dependencies between bundles, this dependency feature
is unique to Porter, and **only bundles that have been authored by Porter can be used as dependencies.**

Here is a [full example][example] of a Porter manifest that uses dependencies.

## Define a dependency

In the manifest, list the dependencies in the order that they should be
installed.

```yaml
dependencies:
  mysql:
    tag: deislabs/porter-mysql:latest
```

Currently dependencies are assumed to be located in one of the locations below.
Once the CNAB spec has a mechanism for distributing bundles via registries, this
will be revisited.

* bundles (in the same directory as the Porter manifest)
* PORTER_HOME/bundles


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

## Defaulting Parameters

Parameters defined in a dependent bundle can be defaulted from the root bundle.
In the example below, the mysql bundle defined the `database_name` and
`mysql_user` parameters, and the root bundle (Wordpress) defaulted those parameters
to a specific value, so that the user wouldn't be required to choose a value for
those parameters when installing Wordpress.

```yaml
dependencies:
  mysql:
    tag: deislabs/porter-mysql:latest
    parameters:
      database_name: wordpress
      mysql_user: wordpress
```

## Dependency Graph

Currently Porter only supports direct dependencies. Dependencies of dependencies, 
transitive dependencies, are ignored. See [Design: Dependency Graph Resolution](https://github.com/deislabs/porter/issues/69) for more information.

## Dependency Ordering

Depending on the action being performed, Porter handles executing the dependent bundle's
steps either before or after the steps in the root bundle.

* Install - The steps are executed _before_ the steps in the root bundle.
* Upgrade - The steps are executed _before_ the steps in the root bundle.
* Uninstall - The steps are executed _after_ the steps in the root bundle.

[example]: https://github.com/deislabs/porter/blob/master/build/testdata/bundles/wordpress/porter.yaml