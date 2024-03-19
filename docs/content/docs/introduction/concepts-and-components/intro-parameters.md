---
title: Parameters
description: The lifecycle of a parameter from definition, to resolution, and finally injection at runtime
weight: 2
aliases:
  - /how-parameters-work/
---

When you are authoring a bundle, you can define parameters that are required by
your bundle. These parameters are restricted to a list of [allowable data
types](/docs/bundle/manifest) and are used to define parameters such as
username and password values for a backing database, or the region that a
certain resource should be deployed in, etc. Then in your action's steps you can
reference the parameters using porter's template language `${
bundle.parameters.db_name }`.

When a bundle is executed, parameter values are resolved in a hierarchy of (from highest to lowest precedence):

- Overrides specified either with the \--param flag or by setting parameters directly on an installation resource
- Parameter sets specified with the \--parameter-set flag or by setting the parameter set name directly on an installation resource
- Parameter values remembered from the last time the bundle was run
- Defaults defined for the parameter in the bundle definition

Let's use the hello-llama bundle from the [Parameters QuickStart](/docs/quickstart/parameters/) as an example and walk through the various ways that Porter will resolve the name parameter.

1. The bundle has one parameter, name, and it has a default therefore we do not need to specify it when installing the bundle. Running `porter install test -r ghcr.io/getporter/hello-llama:v0.1.1` uses the default value of "llama" for the name parameter.
2. Now we can override that name using the \--param flag, `porter upgrade test --param name=sparkles`.
3. When we repeat the upgrade command without specifying the name parameter, the name parameter continues to be "sparkles" as specified the last time the bundle was run. You can see the last used parameter values for an installation with `porter installation show`.
4. You can also override the name using a parameter set and using it during the upgrade, `porter upgrade test --parameter-set hello-llama`.

## Parameter Sets

A Parameter Set is a file which maps parameters to their strategies for value
resolution. Strategies include resolving from a source on the user's machine
such as an environment variable (`env`), filepath (`path`), command result
(`command`) or simply the value itself (`value`). In addition, a parameter
can have a secret source (`secret`). See the [secrets
plugin docs](/plugins/types/#secrets) to learn how to configure Porter to use
an external secret store.

Parameter Sets are created using the combination of [porter parameters create](/docs/references/cli/parameters/create)
and [porter parameters apply](/docs/references/cli/parameters/apply).
Afterwards a parameter set can be [edited](/docs/references/cli/parameters/edit) if changes are required.
See [porter parameters help](/docs/references/cli/parameters/) for all available commands.

Now when you execute the bundle you can pass the name of the parameter set to
the command using the `--parameter-set` or `-p` flag, e.g.
`porter install -p myparamset`.

If you are creating parameter sets manually, you can use the [Parameter Set Schema]
to validate that you have created it properly.

[Parameter Set Schema]: https://github.com/getporter/porter/blob/main/pkg/schema/parameter-set.schema.json

## User-specified values

A user may also supply parameter values when invoking an action on the bundle.
User-supplied values take precedence over both the bundle defaults and any
included in a provided parameter set file. The CLI flag for supplying a
parameter override is `--param`.

For example, you may decide to override the `db_name` parameter for a given
installation via `porter install --param db_name=mydb -p myparamset`.

When a parameter's bundle definition is set to `sensitive=true`, the user-specified
value will be stored into a secret store to prevent security leakage. See the [secrets
plugin docs](/plugins/types/#secrets) to learn how porter uses an external secret store
to handle sensitive data.

## Bundle defaults

The bundle author may have decided to supply a default value for a given
parameter as well. This value would be used when neither a user-specified
value nor a parameter set value is supplied. See the `Parameters` section in
the [Author Bundles](/docs/bundle/manifest) doc for more info.

## Q & A

### Why can't the parameter source be defined in porter.yaml?

See the helpful explanation in the [credentials](/docs/introduction/concepts-and-components/intro-credentials/) doc, which
applies to parameter sources as well.

[create]: /docs/references/cli/parameters/create/
[apply]: /docs/references/cli/parameters/apply/
[edit]: /docs/references/cli/parameters/edit/

## Related

- [QuickStart: Use parameters with a bundle](/docs/quickstart/parameters/)
