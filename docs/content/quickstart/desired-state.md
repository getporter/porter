---
title: "QuickStart: Desired State"
descriptions: Manage an installation by defining its desired state in a file
layout: single
---

You now know how to install a bundle, passing in a credential or parameter set.
[Managing installations] directly with the install and upgrade commands is just one way of working with installations.
Porter also supports something known as **desired state**, where you specify how you want an installation to look in a file, and then Porter handles calling the appropriate commands to match that state.
This QuickStart walks you through how to manage credential sets, parameter sets and installations with the apply commands.

## Explain the bundle

First, let's look at the bundle used in this QuickStart.

```console
$ porter explain ghcr.io/getporter/examples/credentials-tutorial:v0.3.0
Name: examples/credentials-tutorial
Description: An example Porter bundle with credentials. Uses your GitHub token to retrieve your public user profile from GitHub.
Version: 0.3.0
Porter Version: v1.0.0-alpha.19

Credentials:
--------------------------------------------------------------------------------
  Name          Description                          Required  Applies To
--------------------------------------------------------------------------------
  github-token  A GitHub Personal Access             true      install,upgrade
                Token. Generate one at
                https://github.com/settings/tokens.
                No scopes are required.

Parameters:
------------------------------------------------------------------------------------
  Name  Description                     Type    Default  Required  Applies To
------------------------------------------------------------------------------------
  user  A GitHub username. Defaults to  string           false     install,upgrade
        the current user.
```

## Define Credential and Parameter Sets

Instead of creating the credential and parameter sets using the generate command, we will define them in files.
Create a file named creds.yaml, paste the following definition into the file, and then save it.

```yaml
schemaVersion: 1.0.1
name: github
credentials:
  - name: github-token
    source:
      env: GITHUB_TOKEN
```

Create a file named params.yaml, paste the following definition into the file, and then save it.

```yaml
schemaVersion: 1.0.1
name: credentials-tutorial
parameters:
  - name: user
    source:
      value: getporterbot
```

Use the [porter credentials apply] command to import the credential set into Porter:

```
porter credentials apply creds.yaml
```

You can verify that the credential set was imported successfully by viewing its definition with the show command:

```console
$ porter credentials show github
Name: github
Namespace: demo
Created: 0001-01-01
Modified: 7 seconds ago

-------------------------------------------
  Name          Local Source  Source Type
-------------------------------------------
  github-token  GITHUB_TOKEN  env
```

Use the [porter parameters apply] command to import the parameter set into Porter:

```
porter parameters apply params.yaml
```

Verify that the parameter set was imported successfully with the show command:

```console
$ porter parameters show credentials-tutorial
Name: credentials-tutorial
Created: 0001-01-01
Modified: 5 seconds ago

-----------------------------------
  Name  Local Source  Source Type
-----------------------------------
  user  getporterbot  value
```

## Define the installation

Define an installation that uses these parameter and credential sets to install the getporter/credentials-tutorial:0.2.0 bundle.
Create a file named installation.yaml, paste the following definition into the file, and then save it.

```yaml
schemaVersion: 1.0.2
name: desired-state
bundle:
  repository: getporter/credentials-tutorial
  version: 0.3.0
parameterSets:
  - credentials-tutorial
credentialSets:
  - github
```

Use the [porter installation apply] command to import the installation into Porter.
Unlike the other apply commands, this not only imports the definition but also automatically runs either porter install or porter upgrade to make the installation's status match what is defined in the file.
In this case, Porter will run the install command because the installation has not been successfully installed yet.

```console
$ porter installation apply installation.yaml
Created demo/desired-state installation
Triggering because the installation has not completed successfully yet
The installation is out-of-sync, running the install action...
# bundle output truncated for brevity
```

Re-run the apply command and notice that this time the bundle is not executed because the installation is already in the desired state.

```console
$ porter installation apply installation.yaml
Updated demo/desired-state installation
The installation is already up-to-date.
```

## Force

You can force Porter to execute the bundle again with the \--force flag.

```console
$ porter installation apply installation.yaml --force
Updated demo/desired-state installation
The installation is up-to-date but will be re-applied because --force was specified
The installation is out-of-sync, running the install action...
# bundle output truncated for brevity
```

## Modify the Parameter Set

Edit params.yaml, change the user parameter value from getporterbot to carolynvs, and then apply the file.

```yaml
schemaVersion: 1.0.1
name: credentials-tutorial
parameters:
  - name: user
    source:
      value: carolynvs
```

```
porter parameters apply params.yaml
```

This updates the definition of the parameter set, but does not trigger any bundles to execute.
In the future, this behavior may change, but for now you must apply the installation to trigger Porter's reconciliation.
This time Porter will run the upgrade command because the user parameter changed.

```console
$ porter installation apply installation.yaml
Updated demo/desired-state installation
Triggering because the parameters have changed.
Diff:
  map[string]string{
- 	"user": "getporterbot",
+ 	"user": "carolynvs",
  }

The installation is out-of-sync, running the upgrade action...
# bundle output truncated for brevity
```

## Modify the Installation

Now, let's set the user parameter directly on the installation.
Parameters specified on the installation take precedence over parameter values specified in parameter sets.
Edit installation.yaml and add the following lines:

```yaml
parameters:
  user: vdice
```

The installation.yaml file should look like this:

```yaml
schemaVersion: 1.0.0
name: desired-state
bundle:
  repository: getporter/credentials-tutorial
  version: 0.3.0
parameterSets:
  - credentials-tutorial
credentialSets:
  - github
parameters:
  user: vdice
```

Apply the installation.yaml file to trigger an upgrade:

```console
$ porter installation apply installation.yaml
Updated demo/desired-state installation
Triggering because the parameters have changed.
Diff:
  map[string]string{
- 	"user": "carolynvs",
+ 	"user": "vdice",
  }

The installation is out-of-sync, running the upgrade action...
# bundle output truncated for brevity
```

## Dry Run

Sometimes you may want to ask Porter if changes to an installation file _would trigger the bundle to run_, without modifying the installation and potentially triggering it to run again.
The \--dry-run flag checks the definition the installation in the specified file against the current state of the installation, and then tells you what Porter would do.
The installation definition is not modified in Porter's database, nor is the bundle ever run.

Repeat the apply command with the \--dry-run flag, and verify that Porter would not run the bundle:

```console
$ porter installation apply installation.yaml --dry-run
Updated demo/desired-state installation
The installation is already up-to-date.
```

Edit installation.yaml, remove the user parameter, and then save the file.
Repeat the apply command with the \--dry-run flag again, and note that Porter detected the change and would run the upgrade action.

```console
$ porter installation apply installation.yaml --dry-run
Updated demo/desired-state installation
Triggering because the parameters have changed.
Diff:
  map[string]string{
- 	"user": "vdice",
+ 	"user": "carolynvs",
  }

The installation is out-of-sync, running the upgrade action...
Skipping bundle execution because --dry-run was specified
```

## Uninstall

Installations have a field named **uninstalled** that control if the installation should be uninstalled.
After a bundle has been installed, set uninstalled to true on the installation to uninstall it.

Edit installation.yaml, set uninstalled to true, and then save the file.

```yaml
uninstalled: true
```

The installation.yaml file should look like this:

```yaml
schemaVersion: 1.0.0
name: desired-state
uninstalled: true
# remaining fields are not relevant to uninstalling
```

Now, apply the installation.yaml file to trigger an uninstall:

```console
$ porter installation apply installation.yaml
Updated quickstart/desired-state installation
Triggering because installation.uninstalled is true
The installation is out-of-sync, running the uninstall action...
# bundle output truncated for brevity
```

## Next Steps

In this QuickStart you learned how to manage installations using desired state by defining the installation in a file, and then triggering reconciliation of that installation with the apply command.

* [Understand the difference between imperative commands and desired state](/end-users/installations/)
* [Automating Porter with the Porter Operator](/operator/)
* [Create a bundle](/bundle/create/)

[managing installations]: /end-users/installations/
[porter credentials apply]: /cli/porter_credentials_apply/
[porter parameters apply]: /cli/porter_parameters_apply/
[porter installation apply]: /cli/porter_installation_apply/
