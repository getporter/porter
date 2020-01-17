---
title: "porter mixins install"
slug: porter_mixins_install
url: /cli/porter_mixins_install/
---
## porter mixins install

Install a mixin

### Synopsis

Install a mixin

```
porter mixins install NAME [flags]
```

### Examples

```
  porter mixin install helm --url https://cdn.porter.sh/mixins/helm
  porter mixin install helm --feed-url https://cdn.porter.sh/atom.xml
  porter mixin install azure --version v0.4.0-ralpha.1+dubonnet --url https://cdn.porter.sh/mixins/azure
  porter mixin install kubernetes --version canary --url https://cdn.porter.sh/mixins/kubernetes
```

### Options

```
      --feed-url string   URL of an atom feed where the mixin can be downloaded (default https://cdn.porter.sh/atom.xml)
  -h, --help              help for install
      --url string        URL from where the mixin can be downloaded, for example https://github.com/org/proj/releases/downloads
  -v, --version string    The mixin version. This can either be a version number, or a tagged release like 'latest' or 'canary' (default "latest")
```

### Options inherited from parent commands

```
      --debug   Enable debug logging
```

### SEE ALSO

* [porter mixins](/cli/porter_mixins/)	 - Mixin commands. Mixins assist with authoring bundles.

