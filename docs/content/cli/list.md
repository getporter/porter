---
title: "porter list"
slug: porter_list
url: /cli/porter_list/
---
## porter list

List installed bundles

### Synopsis

List all bundles installed by Porter.

A listing of bundles currently installed by Porter will be provided, along with metadata such as creation time, last action, last status, etc.
Optionally filters the results name, which returns all results whose name contain the provided query.
The results may also be filtered by associated labels and the namespace in which the installation is defined. 

Optional output formats include json and yaml.

```
porter list [flags]
```

### Examples

```
  porter list
  porter list -o json
  porter list --label owner=myname --namespace dev
  porter list --name myapp
```

### Options

```
  -h, --help               help for list
  -l, --label strings      Filter the installations by a label formatted as: KEY=VALUE. May be specified multiple times.
      --name string        Filter the installations where the name contains the specified substring.
  -n, --namespace string   Filter the installations by namespace. Defaults to the global namespace.
  -o, --output string      Specify an output format.  Allowed values: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --debug                  Enable debug logging
      --debug-plugins          Enable plugin debug logging
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

