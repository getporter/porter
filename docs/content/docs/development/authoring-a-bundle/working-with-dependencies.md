---
title: Dependencies
description: Dependency Management with Porter
weight: 3
---

In the Porter manifest you can [define a dependency](#define-a-dependency) on another
bundle. The dependent bundle is executed _before_ the bundle is installed, updated, or a custom action is invoked.
The dependent bundle is uninstalled _after_ the root bundle is uninstalled.

Dependencies are an extension of the [CNAB Spec](https://github.com/cnabio/cnab-spec/blob/master/500-CNAB-dependencies.md).
The Dependency specification is still evolving and we are using Porter to act as an initial implementation. So other CNAB
tools may not support dependencies initially.

Here is a [full example][example] of a Porter manifest that uses dependencies.

## Define a dependency

In the manifest, add entries for each dependency of your bundle. The `name` field takes a short name for the dependent bundle that
you will use to reference the dependent bundle elsewhere in the bundle. For example you can reference the dependent bundle's
outputs via `${ bundle.dependencies.NAME.outputs }`. The `reference` field takes the bundle reference of the dependency.
Both `name` and `reference` are required fields.

```yaml
dependencies:
  requires:
    - name: mysql
      bundle:
        reference: getporter/mysql:v0.1.3
```

## Define dependencies v2 (Shared)

DependenciesV2 is under the [**experimental** flag](https://porter.sh/docs/configuration/configuration/#experimental-feature-flags), and therefore shared dependencies
is an experimental feature. Proceed with caution.

To enable DependenciesV2 you must set the experimental flag. This can be done
by setting an environment variable:
```
PORTER_EXPERIMENTAL=dependencies-v2
```

The configuration for DependenciesV2 is similar to that of the first version, except there is now a "sharing" section with the following required fields: `mode`, `group.name`.
`mode` is a boolean, and `group.name` is the identifier which will allow for certain
bundles to share parameters and outputs between each other.

```yaml
dependencies:
  requires:
    - name: mysql
      bundle:
        reference: localhost:5000/mysql:v0.1.0
      sharing:
        mode: true
        group:
          name: myapp
      parameters:
        database-name: wordpress
        mysql-user: wordpress
        namespace: wordpress
```

If there is an existing dependency installed that the parent bundle should connect to, you must create a label for the existing dependency with the `sh.porter.SharingGroup` key, and the value of the group name specified in the parent bundle. The existing dependency **must** be successfully installed. If it is uninstalled this key must be deleted by the users before the operation can proceed. 

```
porter install --label sh.porter.SharingGroup=myapp
```

There are some safeguards in place to make the existing dependency not changed so that it can break other bundles dependening on it, therefore on the following actions this will occur:

**Install**: For parent bundle on existing dependency, the dependency arguments will be passed to the parent. No further changes.

**Upgrade**: The parent bundle will execute the upgrade action, but it will not change anything about the existing dependency.

**Invoke**: Any changes that happen here **will** change the existing dependency. It will be on the user to handle propgating those changes to other parent bundles if needed.

**Uninstall**: The parent bundle will be uninstalled, but the existing dependency will not be and need to be uninstalled in a separate command.


## Ordering of dependencies

If more than one dependency is declared, they will be installed in the order they are listed. For example, if both the `mysql` and
`nginx` bundles are required, but the `mysql` bundle should be installed first, you would list them as such:

```yaml
dependencies:
  requires:
    - name: mysql
      bundle:
        reference: getporter/mysql:v0.1.3
    - name: nginx
      bundle:
        reference: my/nginx-bundle:v0.1.0
```

## Defaulting Parameters

Parameters defined in a dependent bundle can be defaulted from the root bundle.
In the example below, the mysql bundle defines `database_name` and
`mysql_user` parameters, and the root bundle (Wordpress) has chosen to default those parameters
to specific values, so that the user isn't required to provide values for those parameter.

```yaml
dependencies:
  requires:
    - name: mysql
      bundle:
        reference: getporter/mysql:v0.1.3
      parameters:
        database_name: wordpress
        mysql_user: wordpress
```

## Specifying parameters

### Command-line

You can specify parameters for a dependent bundle on the command-line using the following syntax

```
--param DEPENDENCY#PARAMETER=VALUE
```

For example, to override the default parameter `database_name` when installing the wordpress bundle the comand would be

```
$ porter install --reference getporter/mysql:v0.1.3 --param mysql#database_name=mywordpress
```

- `DEPENDENCY`: The dependency name used in the `dependencies` section of the porter manifest. From the example above, the name is "mysql".
- `PARAMETER`: The name of the parameter.
- `VALUE`: The parameter value.

### Parameter Set

The same syntax shown above can be used to specify dependency parameters in a [Parameter Set][parameter-set] file.

Here, the `name` field should be set to the `DEPENDENCY#PARAMETER` value, or `mysql#database-name` from above.

```json
{
  "name": "wordpress",
  "parameters": [
    {
      "name": "mysql#database-name",
      "source": {
        "value": "mywordpress"
      }
    }
  ]
}
```

### Parameter Precedence

A parameter for a dependency can be set in a few places, here is the order of precedence:

1. Parameters set directly on the command-line via `--param`
1. Parameters set in a [Parameter Set][parameter-set] file via `--parameter-set`
1. Parameters set using a dependency default, for example
   ```yaml
   dependencies:
     requires:
       - name: mysql
         bundle:
           reference: getporter/mysql:v0.1.3
         parameters:
           database_name: wordpress
   ```
1. Parameter defaults defined in a bundle, for example
   ```yaml
   parameters:
     - name: database_name
       type: string
       default: mydb
   ```

## Dependency Registry Authentication

Porter support multiple registries. If you are pulling different dependencies from multiple registries,
just make sure that you login into each registry before executing your bundle.

For example:

```shell
  echo $(DOCKER_REGISTRY_TOKEN) | docker login $(DOCKER_REGISTRY_URL) --username $(DOCKER_REGISTRY_USERNAME) --password-stdin
  echo $(PRIVATE_REGISTRY_TOKEN) | docker login $(PRIVATE_REGISTRY_URL) --username $(PRIVATE_REGISTRY_USERNAME) --password-stdin
  echo $(SECOND_PRIVATE_REGISTRY_TOKEN) | docker login $(SECOND_PRIVATE_REGISTRY_URL) --username $(SECOND_PRIVATE_REGISTRY_USERNAME) --password-stdin
```

## Dependency Graph

At this time Porter only supports direct dependencies. Dependencies of dependencies, a.k.a.
transitive dependencies, are ignored. See [Design: Dependency Graph Resolution](https://github.com/getporter/porter/issues/69)
for our backlog item tracking this feature. We do plan to support it!

[example]: /src/build/testdata/bundles/wordpress/porter.yaml
[parameter-set]: /parameters#parameter-sets
