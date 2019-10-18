---
title: "porter inspect"
slug: porter_inspect
url: /cli/porter_inspect/
---
## porter inspect

Inspect a bundle

### Synopsis

Inspect a bundle by printing the parameters, credentials, outputs, actions and images.

```
porter inspect [flags]
```

### Examples

```
  porter inspect
  porter inspect --file another/porter.yaml
  porter inspect --cnab-file some/bundle.json
  porter inspect --tag deislabs/porter-bundle:v0.1.0
		  
```

### Options

```
      --cnab-file string   Path to the CNAB bundle.json file.
  -f, --file porter.yaml   Path to the Porter manifest. Defaults to porter.yaml in the current directory.
  -h, --help               help for inspect
  -o, --output string      Specify an output format.  Allowed values: table, json, yaml (default "table")
  -t, --tag string         Use a bundle in an OCI registry specified by the given tag
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

