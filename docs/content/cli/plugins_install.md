---
title: "porter plugins install"
slug: porter_plugins_install
url: /cli/porter_plugins_install/
---
## porter plugins install

Install a plugin

### Synopsis

Install a plugin

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
      --feed-url string   URL of an atom feed where the plugin can be downloaded (default https://cdn.porter.sh/plugins/atom.xml)
  -h, --help              help for install
      --url string        URL from where the plugin can be downloaded, for example https://github.com/org/proj/releases/downloads
  -v, --version string    The plugin version. This can either be a version number, or a tagged release like 'latest' or 'canary' (default "latest")
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter plugins](/cli/porter_plugins/)	 - Plugin commands. Plugins enable Porter to work on different cloud providers and systems.

