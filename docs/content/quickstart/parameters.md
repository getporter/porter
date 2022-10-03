---
title: "QuickStart: Parameters"
descriptions: Learn how to use a bundle with parameters
layout: single
---

Now that you know how to install a bundle, let's look at how to specify parameters to customize how that bundle is installed.
Bundle authors define parameters in bundles so that end-users can tweak how the bundle is configured at installation.
A parameter can be a string, integer, boolean or even a json object.
Some examples of how parameters can be used in a bundle are:

* Log Level: Default the log level for an application to info. At any time a user can upgrade the bundle to change that parameter to a different value.
* Deployment Region: Let the user specify which region, such as eastus-1, where the application should be deployed.
* Helm Release Name: A bundle that uses Helm will often define a parameter that allows the user to set the release name for the Helm release.

For optional parameters, bundles set a default value that is used when the user does not specify a value for the parameter.

Let's look at a bundle with parameters:

```console
$ porter explain getporter/hello-llama:v0.1.1
Name: hello-llama
Description: An example Porter bundle with parameters
Version: 0.1.0
Porter Version: v0.38.1-32-gb76f5c1c

Parameters:
Name   Description                           Type     Default   Required   Applies To
name   Name of to whom we should say hello   string   llama     false      All Actions

```

In the Parameters section of the output returned by explain, there is a single optional string parameter, name, with a default of "llama" that applies to "All Actions".
This means that the name parameter can be specified with every action that the bundle defines.
In the custom actions section of the output returned by explain, there are no custom actions defined.
The hello-llama bundle only supports the built-in actions of install, upgrade, and uninstall.

## Specifying parameters

Pass parameters to a bundle with the \--param flag, where the flag value is formatted as PARAM=VALUE.
For example:

```
porter install --param name=Robin
```

When trying out a bundle, it might work well to set individual parameter values on the command line with the --param flag.
Parameter sets store multiple parameters and pass them to a bundle using the parameter set name.
With parameter sets you can avoid errors, and the requirement of remembering and manually configuring parameters at the command line.
Parameter sets store the parameter name, and the source of the parameter value which could be a:

* hard-coded value
* environment variable
* file
* command
* secret

Some parameters may be sensitive, for example a database connection string or oauth token.
For improved security, and to limit exposure of sensitive values, it is recommended that you source sensitive parameter values from a secret store such as HashiCorp Vault or Azure Key Vault.
See the list of available [plugins](/plugins/) for which secret providers are supported.

Porter stores all sensitive parameter values in a secret store, never in Porter's database.
Sensitive parameter values are resolved from the secret store just-in-time before the bundle run.

## Use the default parameter values

Install the bundle without specifying any parameters so that you can see the default behavior of the bundle.

```console
$ porter install hello-llama --reference getporter/hello-llama:v0.1.1
installing hello-llama...
executing install action from hello-llama (installation: hello-llama)
Hello, llama
execution completed successfully!
```

The bundle printed "Hello, llama" using the default value for the name parameter.

## Specify a parameter with a flag

Next upgrade the installation and change the name parameter to another value using the \--param flag.

```console
$ porter upgrade hello-llama --param name=Michelle
upgrading hello-llama...
executing upgrade action from hello-llama (installation: hello-llama)
Michelle 2.0
execution completed successfully!
```

## Create a Parameter Set

Create a parameter set for the hello-llama with the combination of `porter parameters create` and `porter parameters apply` commands. The `create` command will generate a [template file](/reference/file-formats#parameter-set). You need to edit the file to include the corresponding parameters needed for the bundle. After modifying the file, the `apply` command will create the parameter set based on the file. 

```console
$ porter parameters create hello-llama.json
creating porter parameter set in the current directory
$ cat hello-llama.json
# modify hello-llama.json with your editor to the content below
{
    "schemaType": "ParameterSet",
    "schemaVersion": "1.0.1",
    "name": "hello-llama",
    "parameters": [
        {
            "name": "name",
            "source": {
                "value": "Porter"
            }
        }
    ]
}
$ porter parameters apply hello-llama.json
Applied /hello-llama parameter se
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

* [QuickStart: Pass credentials to a bundle](/quickstart/credentials/)
* [Understanding how parameters are resolved](/parameters)
