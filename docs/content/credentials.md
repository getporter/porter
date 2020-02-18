---
title: Credentials
description: The lifecycle of a credential from definition, to resolution, and finally injection at runtime
aliases:
- /how-credentials-work/
---

When you are authoring a bundle, you can define what credentials your bundle
requires such as a github token, cloud provider username/password, etc. Then in
your action's steps you can reference the credentials using porter's template
language `{{ bundle.credentials.github_token }}`.

Credentials are injected when a bundle is executed
(install/upgrade/uninstall/invoke). First a user creates a credentials set using
[porter credentials generate][generate]. This is a mapping that tells porter
"given a name of a credential such as `github_token`, where can the value be
found?". Credential values can be resolved from many places, such as environment
variables or local files, or if you are using a [secrets
plugin](/plugins/types/#secrets) they can come from an external secret store.
The generate command walks you through all the credentials used by a bundle and
where the values can be found.

Now when you execute the bundle you can pass the credential set to the command
use `--cred` or `-c` flag, e.g. `porter install --cred github`. Before the
bundle is executed, porter users the credential set's mappings to retrieve the
credential values and then inject them into the bundle's execution environment,
e.g. the docker container, as environment variables.

Inside the bundle's execution environment Porter looks for those environment
variables that represent the credentials and replaces the template placeholders
like `{{ bundle.credentials.github_token }}` with the actual credential value
before executing the step.

Once the bundle finishes executing, the credentials are NOT recorded in the
bundle instance (claim). Parameters are recorded there so that you can view them
later using `porter instances show NAME --output json`.

## Q & A

### Why can't the credential source be defined in porter.yaml?

The source of the credentials is specific to each installation of the bundle. An
author writes the bundle and defines what credentials are needed by the bundle
and where each credential should be put, for example a certain environment
variable.

When a person installs that bundle only they know where that credential's value
should be resolved from. Perhaps they put it in a environment variable named
after the production environment, or in a file under /tmp, or in their team’s
key vault. This is why the author of the bundle can’t guess and put it in
porter.yaml up front.

[generate]: /cli/porter_credentials_generate/