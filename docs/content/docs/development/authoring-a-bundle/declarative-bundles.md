---
title: Declarative Bundles
description: Why porter.yaml is declarative, and what that means for how you author a bundle.
weight: 1
---

porter.yaml is a **declarative** manifest: you describe the steps you want to run and which [mixins] should run them, not imperative control flow like loops, conditionals, or variable assignment.
Each step names a mixin action and its inputs; the mixin is responsible for figuring out how to make that happen, including handling re-runs safely.

- [Why Porter is declarative](#why-porter-is-declarative)
- [When you need imperative logic](#when-you-need-imperative-logic)
- [Not to be confused with desired state](#not-to-be-confused-with-desired-state)

## Why Porter is declarative

Porter is biased against putting imperative logic directly in porter.yaml, and for now we intend to hold that line rather than grow porter.yaml into a scripting language. A few reasons:

- **Tooling.** Porter can generate the Dockerfile, validate the manifest, and detect edge cases and error conditions, because it understands the shape of a declarative step. It cannot do that for arbitrary embedded scripts.
- **Trust and inspection.** A step that calls a mixin has metadata that describes what it does, so a bundle can be reviewed or vetted based on which mixins it uses and how. A step that shells out to an opaque script can do anything.
- **Composability.** Mixins can collect outputs from one step and pass them as arguments to a later step, even across different mixins, e.g. a `helm` step's connection string output feeding a `terraform` step's input variable. That only works because steps describe data flow declaratively.
- **YAML is a bad scripting language.** Embedding bash inside YAML means fighting quoting and escaping rules for two languages at once. Declarative steps sidestep that entirely.

## When you need imperative logic

You will still run into cases that need real imperative logic: branching, loops, or just a shell one-liner. Porter's answer is to push that logic out of porter.yaml and into something purpose-built for it:

- If there's a mixin for what you're doing, use it. Mixins can adapt imperative, non-idempotent command line tools so they behave well when Porter re-runs them.
- If there isn't, [write a mixin][mixin-dev-guide] so the logic is reusable and testable outside of any one bundle.
- Otherwise, use the [exec mixin], but put the actual scripting in a script file next to your porter.yaml and call it from the step, rather than embedding it in YAML. See [Use scripts][exec-mixin-scripts] in the exec mixin best practices for a worked example and the reasoning behind it.

## Not to be confused with desired state

Porter also uses the words "declarative" and "imperative" for a different distinction: how you *manage installations*, e.g. running `porter install`/`porter upgrade` directly (imperative commands) versus `porter installation apply` against a file describing the installation's desired state. That's about installation lifecycle, not about how porter.yaml itself is written. See [Managing Installations] for that topic.

[mixins]: /mixins/
[mixin-dev-guide]: /mixin-dev-guide/
[exec mixin]: /mixins/exec/
[exec-mixin-scripts]: /docs/best-practices/exec-mixin/#use-scripts
[Managing Installations]: /docs/operations/manage-installations/
