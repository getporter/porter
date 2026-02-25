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
  porter list --all-namespaces,
  porter list --label owner=myname --namespace dev
  porter list --name myapp
  porter list --skip 2 --limit 2
```

### Options

```
      --all-namespaces          Include all namespaces in the results.
      --field-selector string   Selector (field query) to filter on, supports '=' (e.g. --field-selector bundle.version=0.2.0,status.action=install). All fields from the json output are supported.
  -h, --help                    help for list
  -l, --label strings           Filter the installations by a label formatted as: KEY=VALUE. May be specified multiple times.
      --limit int               Limit the number of installations by a certain amount. Defaults to 0.
      --name string             Filter the installations where the name contains the specified substring.
  -n, --namespace string        Filter the installations by namespace. Defaults to the global namespace.
  -o, --output string           Specify an output format.  Allowed values: plaintext, json, yaml (default "plaintext")
      --skip int                Skip the number of installations by a certain amount. Defaults to 0.
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter](/cli/porter/)	 - With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://porter.sh/quickstart to learn how to use Porter.


