---
title: "porter mixins feed generate"
slug: porter_mixins_feed_generate
url: /cli/porter_mixins_feed_generate/
---
## porter mixins feed generate

Generate an atom feed from the mixins in a directory

### Synopsis

Generate an atom feed from the mixins in a directory. 

A template is required, providing values for text properties such as the author name, base URLs and other values that cannot be inferred from the mixin file names. You can make a default template by running 'porter mixins feed template'.

The file names of the mixins must follow the naming conventions required of published mixins:

VERSION/MIXIN-GOOS-GOARCH[FILE_EXT]

More than one mixin may be present in the directory, and the directories may be nested a few levels deep, as long as the file path ends with the above naming convention, porter will find and match it. Below is an example directory structure that porter can list to generate a feed:

bin/
└── v1.2.3/
    ├── mymixin-darwin-amd64
    ├── mymixin-linux-amd64
    └── mymixin-windows-amd64.exe

See https://getporter.org/mixin-dev-guide/distribution more details.


```
porter mixins feed generate [flags]
```

### Examples

```
  porter mixin feed generate
  porter mixin feed generate --dir bin --file bin/atom.xml --template porter-atom-template.xml
```

### Options

```
  -d, --dir string        The directory to search for mixin versions to publish in the feed. Defaults to the current directory.
  -f, --file string       The path of the atom feed output by this command. (default "atom.xml")
  -h, --help              help for generate
  -t, --template string   The template atom file used to populate the text fields in the generated feed. (default "atom-template.xml")
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter mixins feed](/cli/porter_mixins_feed/)	 - Feed commands

