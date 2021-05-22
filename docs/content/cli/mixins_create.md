---
title: "porter mixins create"
slug: porter_mixins_create
url: /cli/porter_mixins_create/
---
## porter mixins create

Create a new mixin project based on the getporter/skeletor repository

### Synopsis

Create a new mixin project based on the getporter/skeletor repository.

The first argument is the name of the mixin to create and is required.
A flag of --author to declare the author of the mixin is also a required input.
You can also specify where to put mixin directory. It will default to the current directory.

```
porter mixins create NAME --author YOURNAME [--dir /path/to/mixin/dir] [flags]
```

### Examples

```
 porter mixin create MyMixin --author MyName
		porter mixin create MyMixin --author MyName --dir path/to/mymixin
		
```

### Options

```
      --author string   Name of the mixin's author.
      --dir string      Path to the designated location of the mixin's directory.
  -h, --help            help for create
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter mixins](/cli/porter_mixins/)	 - Mixin commands. Mixins assist with authoring bundles.

