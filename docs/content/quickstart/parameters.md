---
title: "QuickStart: Parameters"
descriptions: Learn how to use a bundle with parameters
layout: single
---

Now that you know how to install a bundle, let's look at how to specify parameters to customize how that bundle is installed.
Bundles can define parameters to allow the end-user to tweak how the bundle is configured.
A parameter can be a string, integer, boolean or even a json object.
Some examples of how parameters can be used in a bundle are:

* Log Level: Default the log level for an application to info. At any time a user can upgrade the bundle to change that parameter to a different value.
* Deployment Region: Let the user specify which region, such as eastus-1, where the application should be deployed.
* Helm Release Name: A bundle that uses Helm often will define a parameter that allows the user to set the release name for the Helm release.

For optional parameters, bundles set a default value that is used when the user does not specify a value for the parameter.

Let's look at a bundle with parameters:

```console
$ porter explain --reference getporter/hello-llama:v0.1.1
Name: hello-llama
Description: An example Porter bundle with parameters
Version: 0.1.0
Porter Version: v0.38.1-32-gb76f5c1c

No credentials defined

Parameters:
Name   Description                           Type     Default   Required   Applies To
name   Name of to whom we should say hello   string   llama     false      All Actions

No outputs defined

No custom actions defined

No dependencies defined
```

The output tells us that the bundle has one optional string parameter, name, which defaults to "llama".
The name parameter applies to "All Actions" meaning that this parameter can be specified with every action that the bundle defines.
Since it says that no custom actions are defined, the bundle only supports the built-in actions of install, upgrade, and uninstall.

## Use the default parameter values


First install the bundle and do not specify any parameters so that you can observe how the bundle installs without any customization.

```console
$ porter install hello-llama --reference getporter/hello-llama:v0.1.1
installing hello-llama...
executing install action from hello-llama (installation: hello-llama)
Hello, llama
execution completed successfully!
```

The bundle printed "Hello, llama" using the default value for the name parameter.

## Specify a parameter with a flag
Next upgrade the installation and change the name parameter to another value.
Parameters are specified with the \--param flag: 

```console
$ porter upgrade hello-llama --param name=Porter
upgrading hello-llama...
executing upgrade action from hello-llama (installation: hello-llama)
Porter 2.0
execution completed successfully!
```

## Create a Parameter Set
Setting a parameter value individually on the command line with the \--param flag works well when trying out a bundle.
When working with a bundle often though, remembering and typing out every parameter values every time is error prone.
Parameter Sets store a set of parameter values to use with a bundle.
Even more powerful, the value of the parameters can come from the following sources:

* hard-coded value
* environment variable
* file
* command output
* secret

Some parameters may be sensitive, for example a database connection string or oauth token.
For improved security, and to limit exposure of sensitive values, it is recommended that you source those parameter values from a secret store such as Hashicorp Vault or Azure Key Vault.

Create a parameter set for the hello-llama with the `porter parameters generate` command. It is an interactive command that walks through setting values for every parameter in the specified bundle.

```console
$ porter parameters generate hello-llama --reference getporter/hello-llama:v0.1.1
Generating new parameter set hello-llama from bundle hello-llama
==> 2 parameters declared for bundle hello-llama

? How would you like to set parameter "name"
   [Use arrows to move, space to select, type to filter]
  secret
> specific value
  environment variable
  file path
  shell command

? Enter the value that will be used to set parameter "name"
  Porter
```

This creates a parameter set named hello-llama.
View the parameter set with the `porter parameters show` command:

```console
$ porter parameters show hello-llama
Name: hello-llama
Created: 4 minutes ago
Modified: 4 minutes ago

-----------------------------------
  Name  Local Source  Source Type
-----------------------------------
  name  Porter       value
```

The output shows that the parameter set has one parameter stored in it, the name parameter, and it will use the hard-coded value "Porter".

## Specify a parameter with a Parameter Set

Now re-run the upgrade command, this time specifying the name of the parameter set instead of individually specifying each parameter value with flags.
Parameter sets are specified with the \--parameter-set or -p flag.

```console
$ porter upgrade hello-llama --parameter-set hello-llama
upgrading hello-llama...
executing upgrade action from hello-llama (installation: hello-llama)
Porter 2.0
execution completed successfully!
```

## Cleanup

To clean up the resources installed from this QuickStart, use the `porter uninstall` command. 

```
porter uninstall hello-llama
```

## Next Steps 

In this QuickStart, you learned how to see the parameters defined on a bundle, their default values, and customize the installation of a bundle by specifying alternate values.

* [Understanding how parameters are resolved](/parameters)
