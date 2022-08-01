---
title: "A more YAML friendly template delimiter"
description: "Porter is changing its template delimiter to ${ } to work better with YAML"
date: "2022-07-26"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://twitter.com/carolynvs"
authorimage: "https://github.com/carolynvs.png"
tags: ["v1"]
---

Porter is getting a different template delimiter in v1!
Previously, the template delimiter used in porter.yaml was {{ }} and we are changing it to ${ } starting in v1.0.0-beta.2.
<!--more-->

## What does this change?

The problem with the old delimiter is that it forced all templates to be wrapped in quotes because curly braces are special characters in YAML.
The quotes made it impossible to pass a boolean or numeric value to a mixin because the wrapping quotes always forced the value to be a string.

```yaml
schemaVersion: 1.0.0
name: a-great-bundle

custom:
  aBool: true
  aNumber: 3.33
  aString: v1.2.3

install:
  - mymixin:
      aBool: "${ bundle.custom.aBool }" # renders as "true"
      aNumber: "${ bundle.custom.aNumber }" # renders as "3.33"
      aString: "${ bundle.custom.aString }" # renders as "v1.2.3"
```

The new delimiter, ${ }, works better when embedded in YAML and doesn't require the use of quotes.
Here is the same bundle with the new delimiter, and we can finally use booleans and numbers! 🎉

```yaml
schemaVersion: 1.0.0
name: a-great-bundle

custom:
  aBool: true
  aNumber: 3.33
  aString: v1.2.3

install:
  - mymixin:
      aBool: ${ bundle.custom.aBool } # renders as true
      aNumber: ${ bundle.custom.aNumber } # renders as 3.33
      aString: ${ bundle.custom.aString } # renders as v1.2.3
```

## What do you need to change?

If you do nothing, your bundles will continue to build.
Porter will continue to support the old delimiter as long as your bundle is using schemaVersion 1.0.0-alpha.1.

If you want to use the improved delimiter, edit your porter.yaml, **set your schemaVersion to 1.0.0 and use ${ } in your templates**.
Optionally, you can also remove the wrapping quotes around the templates.

We hope this is a welcome change and please send us feedback if you run into any trouble updating your templates!
