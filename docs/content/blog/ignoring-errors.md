---
title: "Ignoring Mixin Errors and Idempotent Actions"
description: "Porter now supports ignoring the errors from a mixin. The az mixin takes advantage of this new feature to manage resource groups."
date: "2022-01-07"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://twitter.com/carolynvs"
authorimage: "https://github.com/carolynvs.png"
tags: ["mixins"]
summary: |
    Porter now supports ignoring the errors from a mixin. The az mixin takes advantage of this new feature to manage resource groups.
---

When is an error not really an error?
Questions like this come up regularly when automating deployments.
For example, when I repeat a script that creates a resource, I don't want it to error out because the resource already exists.
A lot of command-line tools weren't designed in a way that works well with bundles.
An installation could fail halfway through and need to be repeated.
This is why we encourage bundle authors to use custom mixins, instead of just the exec mixin, because mixins are designed to work with bundles and give you **idemponent** behavior.
Mixins should do what's necessary to match the desired state in the bundle and handle recoverable errors.

Since we can't have a custom mixin for every tool and situation though, the exec mixin just got smarter.
You can now handle errors directly from the exec mixin without having to fall back to writing scripts.

Let's take a look at a few different ways that you can handle errors from the exec mixin.

## Ignore All Errors

Sometimes you want to run a command but don't really care if it fails.
Now I won't tell you what to do in production, but when debugging a bundle, this can be handy.

The snippet below will run a command, and Porter will ignore any errors returned by the command, and continue executing the next step in the action.

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
In the example below, the made-up thing command returns 2 when the resource already exists.
We can check for that and ignore the error when we try to create a thing that already exists.

You can ignore multiple exit codes, and if any match, then the command's error is ignored.

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
For example, if you are automating terraform, use the [terraform mixin](https://porter.sh/mixins/terraform/).
Mixins are meant to adapt a tool to work well inside a bundle.

I used the new ignore errors capability of the exec mixin's library to create a custom command for the [az mixin](https://porter.sh/mixins/az/).
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

The example mixin, Skeletor, has been updated with an example custom command to help get you started.

## Try it out

Bundle authors, try moving some of that custom error handling logic out of bash scripts and into your exec mixin calls.
Mixin authors, take a look at how the az mixin uses the exec mixin library to add error handling.
You can quickly add the same error handling behavior to your mixin, or create a custom command that handles errors automatically.

Give it a try and let us know how it works for you!
If there is a mixin that you would like to use this new error handling with, let us know, and we can help make that happen more quickly.
