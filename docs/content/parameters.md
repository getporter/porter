---
title: Parameters
description: The lifecycle of a parameter from definition, to resolution, and finally injection at runtime
aliases:
- /how-parameters-work/
---

When you are authoring a bundle, you can define parameters that are required by
your bundle. These parameters are restricted to a list of [allowable data
types](/author-bundles/#parameter-types) and are used to define parameters such as
username and password values for a backing database, or the region that a
certain resource should be deployed in, etc. Then in your action's steps you can
reference the parameters using porter's template language `${
bundle.parameters.db_name }`.

Parameter values are resolved from a combination of supplied parameter set
files, user-specified overrides and defaults defined by the bundle itself.
The resolved values are added to a claim receipt, which is passed in to
the bundle execution environment, e.g. the docker container, when the bundle
action is executed (install/upgrade/uninstall/invoke).

## Parameter Sets

A Parameter Set is a file which maps parameters to their strategies for value
resolution.  Strategies include resolving from a source on the user's machine
such as an environment variable (`env`), filepath (`path`), command result
(`command`) or simply the value itself (`value`).  In addition, a parameter
can have a secret source (`secret`).  See the [secrets
plugin docs](/plugins/types/#secrets) to learn how to configure Porter to use
an external secret store.

Parameter Sets are created using the combination of [porter parameters create][create]
and [porter parameters apply][apply].
Afterwards a parameter set can be [edited][edit] if changes are required.
See [porter parameters help](/cli/porter_parameters/) for all available commands.

Now when you execute the bundle you can pass the name of the parameter set to
the command using the `--parameter-set` or `-p` flag, e.g.
`porter install -p myparamset`.

If you are creating parameter sets manually, you can use the [Parameter Set Schema]
to validate that you have created it properly.

[Parameter Set Schema]: /src/pkg/schema/parameter-set.schema.json

## User-specified values

A user may also supply parameter values when invoking an action on the bundle.
User-supplied values take precedence over both the bundle defaults and any
included in a provided parameter set file.  The CLI flag for supplying a
parameter override is `--param`.

For example, you may decide to override the `db_name` parameter for a given
installation via `porter install --param db_name=mydb -p myparamset`.

When a parameter's bundle definition is set to `sensitive=true`, the user-specified
value will be stored into a secret store to prevent security leakage. See the [secrets
plugin docs](/plugins/types/#secrets) to learn how porter uses an external secret store 
to handle sensitive data.


## Bundle defaults

The bundle author may have decided to supply a default value for a given
parameter as well.  This value would be used when neither a user-specified
value nor a parameter set value is supplied.  See the `Parameters` section in
the [Author Bundles](/author-bundles#parameters/) doc for more info.

## Q & A

### Why can't the parameter source be defined in porter.yaml?

See the helpful explanation in the [credentials](/credentials/) doc, which
applies to parameter sources as well.

[create]: /cli/porter_parameters_create/
[apply]: /cli/porter_parameters_apply/
[edit]: /cli/porter_parameters_edit/

## Related

* [QuickStart: Use parameters with a bundle](/quickstart/parameters/)
