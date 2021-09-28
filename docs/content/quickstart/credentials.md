---
title: "QuickStart: Credentials"
descriptions: Learn how to use a bundle with credentials
layout: single
---

Now that you know how to customize a bundle installation with parameters, let's look at how your bundle can authenticate with **credentials**.
Credentials are sensitive values associated with your _identity_ and they are treated different from parameters by Porter.
All parameters, sensitive or otherwise, are stored in the installation history as a record of the values used when a bundle was run.
Credentials are never stored by Porter because it isn't safe to assume that identifying information contained in the credentials can be reused.

Examples of a credential would be: a GitHub Personal Access Token, your cloud provider credentials, or a Kubernetes kubeconfig file.
We classify those values as credentials so that when different people execute that bundle, they provide their own personal credentials and execute the bundle with their user permissions.
In contrast, a database connection string used by your application is considered only to be a sensitive parameter because regardless of who is installing the bundle, the same connection string should be used.

**If you want to use different values depending on the person executing the bundle, use credentials. Otherwise, use sensitive parameters.**

This is a convention recommended by Porter to avoid a situation where Sally installs a bundle with her personal credentials, and then every time another user subsequently upgrades the bundle, her credentials are re-used, making it look like Sally ran the upgrades.
Ultimately the difference between the parameters and credentials is that credentials are never stored or reused by a bundle.

Credentials are injected into a bundle as either an environment variable or a file.
Depending on the bundle, a credential can apply to all actions (install/upgrade/uninstall) or may only apply to a particular action.

Let's look at a bundle with credentials:

```console
$ porter explain --reference getporter/credentials-tutorial:v0.1.0
Name: credentials-tutorial
Description: An example Porter bundle with credentials. Uses your GitHub token to retrieve your public user profile from GitHub.
Version: 0.1.0
Porter Version: v0.38.1

Credentials:
Name           Description                                                                                                   Required   Applies To
github-token   A GitHub Personal Access Token. Generate one at https://github.com/settings/tokens. No scopes are required.   true       install,upgrade
```

In the Credentials section of the output returned by explain, there is a single required credential, github-token, that applies to the install and upgrade actions.
This means that the github-token credential is required to run porter install or porter upgrade, but is not required for porter uninstall.

## Create a Credential Set

Create a credential set for the credentials-tutorial bundle with the `porter credentials generate` command. It is an interactive command that walks through setting values for every credential in the specified bundle.

```console
$ porter credentials generate github --reference getporter/credentials-tutorial:v0.1.0
Generating new credential github from bundle credentials-tutorial
==> 1 credentials required for bundle credentials-tutorial
? How would you like to set credential "github-token"
  [Use arrows to move, space to select, type to filter]
  secret
  specific value
> environment variable
  file path
  shell command
? Enter the environment variable that will be used to set credential "github-token"
  GITHUB_TOKEN
```

This creates a credential set named github.
View the credential set with the `porter credentials show` command:

```console
$ porter credentials show github
Name: github
Created: 21 minutes ago
Modified: 21 minutes ago

-------------------------------------------
  Name          Local Source  Source Type
-------------------------------------------
  github-token  GITHUB_TOKEN  env
```

The output shows that the credential set has one credential defined: github-token. The credential's value is not stored in the credential set, instead it only stores a mapping from the credential name to a location where the credential can be resolved, in this case an environment variable named GITHUB_TOKEN.

In production, it is a best practice to source sensitive values, either parameters or credentials, from a secret store, such as Hashicorp Vault or Azure Key Vault.
Avoid storing sensitive values in files or environment variables on developer and CI machines which could be compromised.
See the list of available [plugins](/plugins/) for which secret providers are supported.

## Specify a credential with a Credential Set

Pass credentials to a bundle with the \--cred or -c flag, where the flag value is either the name of a credential set stored in Porter, or a path to a credential set file.
For example:

```
porter install --cred github
```

The output of this example bundle prints data from your public GitHub user profile.

```plaintext
executing install action from credentials-tutorial (installation: credentials-tutorial)
Retrieve current user profile from GitHub
{
  "login": "carolynvs",
  "id": 1368985,
  "node_id": "MDQ6VXNlcjEzNjg5ODU=",
  "avatar_url": "https://avatars.githubusercontent.com/u/1368985?v=4",
  "gravatar_id": "",
  "url": "https://api.github.com/users/carolynvs",
  "html_url": "https://github.com/carolynvs",
  "followers_url": "https://api.github.com/users/carolynvs/followers",
  "following_url": "https://api.github.com/users/carolynvs/following{/other_user}",
  "gists_url": "https://api.github.com/users/carolynvs/gists{/gist_id}",
  "starred_url": "https://api.github.com/users/carolynvs/starred{/owner}{/repo}",
  "subscriptions_url": "https://api.github.com/users/carolynvs/subscriptions",
  "organizations_url": "https://api.github.com/users/carolynvs/orgs",
  "repos_url": "https://api.github.com/users/carolynvs/repos",
  "events_url": "https://api.github.com/users/carolynvs/events{/privacy}",
  "received_events_url": "https://api.github.com/users/carolynvs/received_events",
  "type": "User",
  "site_admin": false,
  "name": "Carolyn Van Slyck",
  "company": "@Azure ",
  "blog": "carolynvanslyck.com",
  "location": "Chicago, IL",
  "email": null,
  "hireable": null,
  "bio": "Professional Yak Shaver",
  "twitter_username": "carolynvs",
  "public_repos": 244,
  "public_gists": 26,
  "followers": 297,
  "following": 1,
  "created_at": "2012-01-22T21:34:25Z",
  "updated_at": "2021-06-28T14:32:27Z"
}
```

Sometimes it is easier to store the credential set in a file instead of having Porter configured with a storage plugin, for example on a test server.
In that case, you can save the credential set to a file with the following command:

```
porter credentials show github --output json > github-creds.json
```

The contents of the file are shown below:

```json
{
  "schemaVersion": "1.0.0",
  "name": "github",
  "namespace": "",
  "created": "2021-06-29T09:44:32.16657-05:00",
  "modified": "2021-06-29T09:44:32.16657-05:00",
  "credentials": [
    {
      "name": "github-token",
      "source": {
        "env": "GITHUB_TOKEN"
      }
    }
  ]
}
```

Below is an example of specifying the credential set with a filepath:

```
porter install --cred ./github-creds.json
```

## Cleanup

To clean up the resources installed from this QuickStart, use the `porter uninstall` command. 

```
porter uninstall credentials-tutorial
```

## Next Steps 

In this QuickStart, you learned how to see the credentials defined on a bundle, generate a credential set telling Porter where to find the credentials values, and pass credentials when executing a bundle.

* [Understanding how credentials are resolved](/credentials/)
* [Manage an installation using desired state](/quickstart/desired-state/)