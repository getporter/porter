---
title: Credentials
description: The lifecycle of a credential from definition, to resolution, and finally injection at runtime
aliases:
- /how-credentials-work/
---

When you are authoring a bundle, you can define what credentials your bundle
requires such as a github token, cloud provider username/password, etc. Then in
your action's steps you can reference the credentials using porter's template
language `{{ bundle.credentials.github_token }}`, or directly access the 
environment variable or path where the credential is stored.

In the example below, the bundle defines two credentials. A kubeconfig file,
which once passed to the bundle is stored at /root/.kube/config, and a GitHub
token which once passed to the bundle is stored in the GITHUB_TOKEN environment
variable.

```yaml
credentials:
- name: kubeconfig
  path: /root/.kube/config
- name: token
  env: GITHUB_TOKEN
```

The paths and environment variable names used in the credential
declaration represent where the value of the injected credentials are stored
_in the bundle when it is executing_. They are not used to locate the credential,
that is the responsibility of credential sets.

## Credential Sets

Before running a bundle the user must first create a credential set using
[porter credentials generate][generate]. A credentials set is a mapping that tells porter
"given a name of a credential such as `github_token`, where can the value be
found?". Credential values can be resolved from many places, such as environment
variables or local files, or if you are using a [secrets
plugin](/plugins/types/#secrets) they can come from an external secret store.
The generate command walks you through all the credentials used by a bundle and
where the values can be found.

If you are creating credential sets manually, you can use the [Credential Set Schema]
to validate that you have created it properly.

[Credential Set Schema]: /src/pkg/schema/credential-set.schema.json

## Runtime

Now when you execute the bundle you can pass the credential set to the command
with `--cred` or `-c` flags. For example, `porter install --cred github`. Before the
bundle is executed, Porter users the credential set's mappings to retrieve the
credential values, and then injects them into the bundle's execution environment 
as either environment variables or files.

Inside the bundle's execution environment Porter replaces the template placeholders
like `{{ bundle.credentials.github_token }}` with the actual credential value
before executing the step. Credentials are also available directly through the
environment variable or path used in its declaration.

Once the bundle finishes executing, the credentials are NOT recorded in the
installation history. Parameters are recorded there so that you can view them
later using `porter installations show NAME --output json`.

## Q & A

### Can I pass credentials to a bundle without credential sets?

No, credentials must be passed to a bundle using credential sets.
Credentials are sensitive values and should ideally be sourced from a secret store such as Hashicorp Vault or Azure Key Vault to limit their exposure.

If circumstances prevent you from using credential sets stored by Porter, you can export a credential set to a file and pass the file to a bundle as demonstrated below.

```
porter credentials show NAME --output json > creds.json
porter install --cred ./creds.json
```

### Why can't the credential source be defined in porter.yaml?

The source of a credential is specific to each installation of the bundle. An
author writes the bundle and defines what credentials are needed by the bundle
and where each credential should be put, for example a certain environment
variable.

When a person installs that bundle only they know where that credential's value
should be resolved from. Perhaps they put it in a environment variable named
after the production environment, or in a file under /tmp, or in their team’s
key vault. This is why the author of the bundle can’t guess and put it in
porter.yaml up front.

[generate]: /cli/porter_credentials_generate/

## Related

* [QuickStart: Pass credentials to a bundle](/quickstart/credentials/)
