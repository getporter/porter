---
title: "porter"
slug: porter
url: /cli/porter/
---
## porter

With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://getporter.org/quickstart to learn how to use Porter.


```
porter [flags]
```

### Examples

```
  porter create
  porter build
  porter install
  porter uninstall
```

### Options

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
  -h, --help                   help for porter
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
  -v, --version                Print the application version
```

### SEE ALSO

* [porter archive](/cli/porter_archive/)	 - Archive a bundle from a reference
* [porter build](/cli/porter_build/)	 - Build a bundle
* [porter bundles](/cli/porter_bundles/)	 - Bundle commands
* [porter completion](/cli/porter_completion/)	 - Generate completion script
* [porter copy](/cli/porter_copy/)	 - Copy a bundle
* [porter create](/cli/porter_create/)	 - Create a bundle
* [porter credentials](/cli/porter_credentials/)	 - Credentials commands
* [porter explain](/cli/porter_explain/)	 - Explain a bundle
* [porter inspect](/cli/porter_inspect/)	 - Inspect a bundle
* [porter install](/cli/porter_install/)	 - Create a new installation of a bundle
* [porter installations](/cli/porter_installations/)	 - Installation commands
* [porter invoke](/cli/porter_invoke/)	 - Invoke a custom action on an installation
* [porter lint](/cli/porter_lint/)	 - Lint a bundle
* [porter list](/cli/porter_list/)	 - List installed bundles
* [porter logs](/cli/porter_logs/)	 - Show the logs from an installation
* [porter mixins](/cli/porter_mixins/)	 - Mixin commands. Mixins assist with authoring bundles.
* [porter parameters](/cli/porter_parameters/)	 - Parameter set commands
* [porter plugins](/cli/porter_plugins/)	 - Plugin commands. Plugins enable Porter to work on different cloud providers and systems.
* [porter publish](/cli/porter_publish/)	 - Publish a bundle
* [porter schema](/cli/porter_schema/)	 - Print the JSON schema for the Porter manifest
* [porter show](/cli/porter_show/)	 - Show an installation of a bundle
* [porter storage](/cli/porter_storage/)	 - Manage data stored by Porter
* [porter uninstall](/cli/porter_uninstall/)	 - Uninstall an installation
* [porter upgrade](/cli/porter_upgrade/)	 - Upgrade an installation
* [porter version](/cli/porter_version/)	 - Print the application version

