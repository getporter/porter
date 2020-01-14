---
title: "porter bundles inspect"
slug: porter_bundles_inspect
url: /cli/porter_bundles_inspect/
---
## porter bundles inspect

Inspect a bundle

### Synopsis

Inspect a bundle by printing the invocation images and any related images images.

If you would like more information about the bundle, the porter explain command will provide additional information,
like parameters, credentials, outputs and custom actions available.


```
porter bundles inspect [flags]
```

### Examples

```
  porter bundle inspect
  porter bundle inspect --file another/porter.yaml
  porter bundle inspect --cnab-file some/bundle.json
  porter bundle inspect --tag getporter/hello:v0.1.0
		  
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

* [porter bundles](/cli/porter_bundles/)	 - Bundle commands

