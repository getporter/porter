---
title: "porter mixins install"
slug: porter_mixins_install
url: /cli/porter_mixins_install/
---
## porter mixins install

Install a mixin

### Synopsis

Install a mixin.

By default mixins are downloaded from the official Porter mixin feed at https://cdn.porter.sh/mixins/atom.xml. To download from a mirror, set the environment variable PORTER_MIRROR, or mirror in the Porter config file, with the value to replace https://cdn.porter.sh with.

```
porter mixins install NAME [flags]
```

### Examples

```
  porter mixin install helm3 --feed-url https://mchorfa.github.io/porter-helm3/atom.xml
  porter mixin install azure --version v0.4.0-ralpha.1+dubonnet --url https://cdn.porter.sh/mixins/azure
  porter mixin install kubernetes --version canary --url https://cdn.porter.sh/mixins/kubernetes
```

### Options

```
      --feed-url string   URL of an atom feed where the mixin can be downloaded. Defaults to the official Porter mixin feed.
  -h, --help              help for install
      --mirror string     Mirror of official Porter assets (default "https://cdn.porter.sh")
      --url string        URL from where the mixin can be downloaded, for example https://github.com/org/proj/releases/downloads
  -v, --version string    The mixin version. This can either be a version number, or a tagged release like 'latest' or 'canary' (default "latest")
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter mixins](/cli/porter_mixins/)	 - Mixin commands. Mixins assist with authoring bundles.

