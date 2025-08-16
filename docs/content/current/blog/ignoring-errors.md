---
title: "Ignoring Mixin Errors and Idempotent Actions"
description: "Porter now supports ignoring the errors from a mixin. The az mixin takes advantage of this new feature to manage resource groups."
date: "2022-01-25"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://twitter.com/carolynvs"
authorimage: "https://github.com/carolynvs.png"
tags: ["mixins"]
summary: |
    Porter now supports ignoring the errors from a mixin. The az mixin takes advantage of this new feature to manage resource groups.
---

Previously, if you needed to handle an error in a bundle, you had to switch from Porter's declarative mixin syntax to a bash script.
But not anymore! The exec mixin has just added support for determining if the error returned by a command is fatal, and should stop the bundle, or if the error can be ignored. 

A common scenario for ignoring errors is during the install action.
When an install fails halfway through, the user will try again by repeating the install command.
If a resource was created during the first install, repeating a command that creates a resource could result in an error because the resource already exists.
Some command-line tools handle this gracefully, while others return an error.
The exec mixin now lets you decide if that error is fatal based on the command's return code or output.

Let's take a look at a few different ways that you can handle errors from the exec mixin.

## Ignore All Errors

You can ignore all errors that are returned by a command, and continue executing the next step in the action.
This can be useful in a couple scenarios such as debugging while writing a bundle, when a command always returns a non-zero exit code, or when you want to try calling a command and keep going regardless of whether it worked.

```yaml
install:
  - exec:
      description: "This may not work but such is life"
      command: ./buy-me-coffee.sh
      ignoreError:
        all: true # Ignore any errors from this command
```

## Ignore Exit Codes

Sometimes you get lucky, and the command that you are using has well-defined exit codes.
For the following example, the behavior of thing responds with the exit code [2] when the resource already exists.

```yaml
install:
  - exec:
      description: "Ensure thing exists"
      command: thing
      arguments:
        - create
      ignoreError:
        exitCodes: [2]
```

In this example, since you don't care about the implied error as the desired outcome is the existence of resource, you configure porter to ignore the [2] exit code.

You can ignore multiple exit codes, and if any match, then the command's error is ignored.

## Ignore Output Containing a String

Usually we aren't so lucky, and we have to scrape the contents of standard error to figure out what went wrong.
Continuing our efforts to create idempotent resources, we can ignore the error when it contains "thing already exits".

```yaml
install:
  - exec:
      description: "Ensure thing exists"
      command: thing
      arguments:
        - create
      ignoreError:
        output:
          contains: ["thing already exists"]
```

## Ignore Output Matching a Regular Expression

Finally, there are times when the error message is a bit more difficult to parse, so we fall back to our favorite hammer: regular expressions.
In the example below, when we delete a thing that has already been deleted, "thing NAME not found" is printed to standard error.

Regular expressions being the tricky devils they are, I recommend using [regex101.com](https://regex101.com/) to quickly test and iterate on your regular expression.

```yaml
uninstall:
  - exec:
      description: "Make the thing go away"
      command: thing
      arguments:
        - remove
      ignoreError:
        output:
          regex: "thing (.*) not found"
```

## Create Custom Idempotent Mixin Commands

Whenever possible, I encourage you to avoid the exec mixin and use a custom mixin for the tooling that you are automating.
For example, if you are automating terraform, use the [terraform mixin](/mixins/terraform/).
Mixins are meant to adapt a tool to work well inside a bundle.

I used the new ignore errors capability of the exec mixin's library to create a custom command for the [az mixin](/mixins/az/).
The **group** command allows you to declare a resource group, and the mixin will handle creating it if it doesn't exist, and cleaning it up when the bundle is uninstalled.

```yaml
install:
  - az:
      description: "Ensure my resource group exists"
      group:
        name: mygroup
        location: eastus2

uninstall:
  - az:
      description: "Remove my resource group"
      group:
        name: mygroup
```

These mixin commands are idempotent and handle errors automatically for you.
This lets you focus on the resources you need, and spend less time figuring out how to automate a command-line tool to work in a way it wasn't designed for.

## Try it out

Bundle authors, try moving some of that custom error handling logic out of bash scripts and into your exec mixin calls.
Mixin authors, take a look at how the [Skeletor] mixin template source [uses the get.porter.sh/porter/pkg/exec/builder package to include error handling](https://github.com/getporter/skeletor/blob/6261f95d7583d581a778d755612827d7d979e40e/pkg/skeletor/action.go#L112-L115).
You can quickly add the same error handling behavior to your mixin, or create a custom command that handles errors automatically by looking at the [source for the az group command](https://github.com/getporter/az-mixin/blob/v0.6.0/pkg/az/group.go).

Give it a try and let us know how it works for you!
If there is a mixin that you would like to use this new error handling with, let us know, and we can help make that happen more quickly.

[Skeletor]: https://github.com/getporter/skeletor
