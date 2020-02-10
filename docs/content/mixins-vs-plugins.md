---
title: Mixins vs. Plugins
description: What is the difference between a Porter mixin and a plugin? When would you use one instead of another?
---

[Mixins](/mixins/) are the building blocks for authoring bundles. They
understand both Porter and CNAB, so you don't have to, and ideally have built-in
logic to make your bundles easier to author and more robust.  They are involved
when the bundle is built, assisting with generating the Dockerfile for the
invocation image. Most importantly they are responsible for handling the actions
in your porter.yaml when your bundle is run. For example, the helm mixin
installs the helm CLI into the invocation image, and then uses it to execute any
helm actions defined in the bundle.

I like to think of mixins as paint colors on your paint palette 🎨 that you use
to **create** your bundle. Like paints, mixins are **composable** and let you
build something new, more than the sum of its parts, limited only by your
imagination.

[Plugins](/plugins/) **extend** the Porter client itself, **reimplementing**
Porter's default functionality. There are fixed extension points in Porter with
a defined interface. For example, Porter saves claims and credential sets using the local
filesystem to ~/.porter by default. A plugin can change that behavior to save
them to cloud storage instead.

What both mixins and plugins have in common is that anyone can create their own
and distribute them, just like the ones that we install with Porter by default.
The Porter team is committed to making our plugin and mixin ecosystem a level
playing field with a low barrier to entry. Check out the [Mixin Developer
Guide](/mixin-dev-guide) to get started.
