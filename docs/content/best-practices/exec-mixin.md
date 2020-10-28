---
title: Best practices for the exec mixin
description: How to effectively use the exec mixin and avoid common mistakes.
---

The [exec mixin](/mixins/exec/) is intended as the mixin of last resort. 😅 We
say that with love, and an awareness that having given you the exec mixin, we
have also given you a way to use Porter in ways that undermine a lot of the good
stuff that make Porter worth using in the first place.

Here are some best practices to consider when you find yourself using the exec
mixin to make sure that you aren't inadvertently making your life harder than it
needs to be:

* [Use purpose-built mixins](#use-purpose-built-mixins) 
* [Use script files](#use-scripts)
* [Quoting, Escaping, Bash and YAML](#quoting-escaping-bash-and-yaml)

## Use purpose-built mixins

When there is a mixin for the task you are doing, use that mixin instead. Those
mixins have advantages over using the exec mixin:

* Automatic generation of the Dockerfile.
* Detection of edge cases and error conditions at runtime.
* Declarative syntax that provides enhanced metadata when people inspect your
  bundles. A bundle using exec can do anything, but specific mixins can be
  vetted and trusted.

Do you have an idea for a mixin? You can [create your own](/mixin-dev-guide/) or
[suggest one][new-issue] to the community and see if someone is interested in
implementing it.

[new-issue]: https://github.com/getporter/porter/issues/new

## Use scripts

When you do need to call a scripting language, like bash, do not embed it into
porter.yaml. Make functions in a script file next to your porter.yaml and call
those functions from the exec mixin. _We cannot recommend this strongly enough_.
In fact we check for it during `porter build` and will warn when we detect such
shenanigans.

A few reasons why we are being annoying about this:

* Scripts are easier to test, lint and validate.
* Properly escaping code in YAML is just awful. If you open a bug about it, we 
  won't help and instead will tell you to put it in a script. #SorryNotSorry
* It is much easier to read and understand intent.

Here's an example of what we think works well, especially when you need to call
bash one-liners occasionally in your porter.yaml.

See [exec outputs][exec-outputs] for a full working example.

```bash
#!/usr/bin/env bash

generate-users() {
    echo '{"user": "sally"}' > users.json
}

# ... define more functions here

# Call the requested function and pass the arguments as-is
"$@"
```

```yaml
install:
- exec:
    description: "Create a file"
    command: ./cluster.sh
    arguments:
    - generate-users
```

[exec-outputs]: https://porter.sh/src/examples/exec-outputs/

## Quoting, Escaping, Bash and YAML

If you got past the previous section and are still ignoring my good advice 😬,
here are a couple thoughts on where you probably went wrong:

* `bash -c` accepts a quoted string, e.g. `bash -c "your command here"`. When you
  embed that in YAML, you have to embed your quotes in more quotes. Are we having
  fun yet?

    **Broken Example**
    
    This is everyone's first try embedding bash in porter.yaml. Alas it doesn't work...
    ```yaml
    exec:
      description: quotes are for suckers
      command: bash
      flags:
        c: echo "Hello World"
    ```

    **Working Example**

    Why are there two sets of quotes? The outer set is just YAML syntax, the inner quotes are what is captured and used as the command, `bash -c "echo Hello World"`.
    ```yaml
    exec:
      description: such quotes, much wow
      command: bash
      flags:
        c: '"echo Hello World"'
    ```

* YAML strings have different modes for when they will escape special
  characters, such as `\n`. When you aren't having luck, try switching the
  [string deliminator](https://yaml-multiline.info/) you are using. Crying then
  going back to a script file works great too. 😉

    **Broken Example**

    It's easy to forget that single quotes in YAML prevent escapes `\` from working...

    ```yaml
    exec:
      description: quotes are for suckers
      command: bash
      flags:
        c: 'printf "Hello World \t\n"'
    ```

    **Working Example**

    This ugly thing is something you never want to try to craft yourself OR debug. Just please use a script.
    ```yaml
    exec:
      description: never put this into your yaml
      command: bash
      flags:
        c: |+
           'printf "Hello World \t
           "'
    ```
