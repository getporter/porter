---
title: "porter explain"
slug: porter_explain
url: /cli/porter_explain/
---
## porter explain

Explain a bundle

### Synopsis

Explain how to use a bundle by printing the parameters, credentials, outputs, actions.

```
porter explain [flags]
```

### Examples

```
  porter explain
  porter explain --file another/porter.yaml
  porter explain --cnab-file some/bundle.json
  porter explain --tag getporter/porter-hello:v0.1.0
		  
```

### Options

```
      --cnab-file string   Path to the CNAB bundle.json file.
  -f, --file porter.yaml   Path to the Porter manifest. Defaults to porter.yaml in the current directory.
  -h, --help               help for explain
  -o, --output string      Specify an output format.  Allowed values: table, json, yaml (default "table")
  -t, --tag string         Use a bundle in an OCI registry specified by the given tag
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

