---
title: "porter plugins install"
slug: porter_plugins_install
url: /cli/porter_plugins_install/
---
## porter plugins install

Install plugins

### Synopsis


Porter offers two ways to install plugins. Users can install plugins one at a time or multiple plugins through a plugins definition file.

Below command will install one plugin:

porter plugins install NAME [flags]

To install multiple plugins at once, users can pass a file to the install command through --file flag:

porter plugins install --file plugins.yaml

The file format for the plugins.yaml can be found here: https://porter.sh/reference/file-formats/#plugins

By default plugins are downloaded from the official Porter plugin feed at https://cdn.porter.sh/plugins/atom.xml. To download from a mirror, set the environment variable PORTER_MIRROR, or mirror in the Porter config file, with the value to replace https://cdn.porter.sh with.

```
porter plugins install NAME [flags]
```

### Examples

```
  porter plugin install azure  
  porter plugin install azure --url https://cdn.porter.sh/plugins/azure
  porter plugin install azure --feed-url https://cdn.porter.sh/plugins/atom.xml
  porter plugin install azure --version v0.8.2-beta.1
  porter plugin install azure --version canary 
  porter plugin install --file plugins.yaml --feed-url https://cdn.porter.sh/plugins/atom.xml
  porter plugin install --file plugins.yaml --mirror https://cdn.porter.sh
```

### Options

```
      --feed-url string   URL of an atom feed where the plugin can be downloaded. Defaults to the official Porter plugin feed.
  -f, --file string       Path to porter plugins config file.
  -h, --help              help for install
      --mirror string     Mirror of official Porter assets (default "https://cdn.porter.sh")
      --url string        URL from where the plugin can be downloaded, for example https://github.com/org/proj/releases/downloads
  -v, --version string    The plugin version. This can either be a version number, or a tagged release like 'latest' or 'canary' (default "latest")
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter plugins](/cli/porter_plugins/)	 - Plugin commands. Plugins enable Porter to work on different cloud providers and systems.

