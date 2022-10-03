---
title: "porter plugins install"
slug: porter_plugins_install
url: /cli/porter_plugins_install/
---
## porter plugins install

Install a plugin

### Synopsis

Install a plugin.

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
```

### Options

```
      --feed-url string   URL of an atom feed where the plugin can be downloaded. Defaults to the official Porter plugin feed.
  -h, --help              help for install
      --mirror string     Mirror of official Porter assets (default "https://cdn.porter.sh")
      --url string        URL from where the plugin can be downloaded, for example https://github.com/org/proj/releases/downloads
  -v, --version string    The plugin version. This can either be a version number, or a tagged release like 'latest' or 'canary' (default "latest")
```

### Options inherited from parent commands

```
      --experimental strings   Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter plugins](/cli/porter_plugins/)	 - Plugin commands. Plugins enable Porter to work on different cloud providers and systems.

