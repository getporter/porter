---
title: Mixin Architecture 
description: How do mixins work? Hotwiring a porter mixin
---

## What is a Mixin

Porter provides two major capabilities: a bundle building capability and with a bundle run time capability.

A cloud native application bundle built with Porter consists of two main components: the Porter runtime and a declarative manifest file named `porter.yaml`.

The Porter runtime provides the entry point for the bundle and is responsible for executing the desired functionality that has been expressed in the `porter.yaml`. The `porter.yaml` declares what should happen for bundle action. Each bundle action is defined as one or more operations, or steps. Each step definition defines a discrete piece of functionality, such as installing a Helm chart or creating a service in a cloud provider. The Porter runtime is responsible for executing each step, but does not implement the desired functionality itself. The functionality declared in each step is actually provided by components called mixins. Porter invokes the mixin by passing the relevant section of the `porter.yaml` to the mixin via standard input. The mixin accepts the YAML document that describes the desired result and performs the desired action. Once finished, the mixin will write any desired outputs to standard out and return control to the Porter runtime.

For example, a bundle author may wish to execute a bash command. The author would include the `exec` mixin in the mixins section of the porter.yaml and then create a step that defines the desired bash command. The following bundle example would execute `bash -c "echo Hello World"`:

```yaml
mixins:
- exec

name: hello
version: 0.1.0
regisry: getporter

install:
- exec:
    description: "Say Hello"
    command: bash
    flags:
      c: echo Hello World
```

## Mixin API

Porter defines a contract that mixins must fulfill in order to be included in the Porter ecosystem. This contract specifies how Porter will execute the mixins, as show above, but also specifies how the mixin is used to build the invocation image for the bundle. Additionally, the contract specifies how a mixin can specify the inputs it can accept and the outputs that it can provide.

Here's a diagram that illustrates how mixins fits into Porter's execution flow:
<img src="/images/mixins/flow-chart.png" style="max-width: 80%; height: auto;"/>

### Build Time

The previous example introduced how mixins are used when a bundle is executed. In the case of the `exec` mixin, the resulting bundle invocation image already has everything needed to execute the mixin. Other mixins may require additional runtime software. The `helm` mixin, for example, requires the [Helm](https://helm.sh/) client at runtime. Porter is responsible for building the invocation image for the bundle, so it needs to know what each mixin will need so that it can be included in the bundle. The mixin is responsible for providing any relevant lines to ensure that the generated invocation image Dockerfile has all required runtime components. Porter expects that the mixin will provide any relevant Dockerfile additions through a `build` command. The `build` should output any necessary Dockerfile commands to standard out. To see this in action, consider the `helm` mixin:

```console
$ ./bin/mixins/helm/helm build
RUN apt-get update && \
 apt-get install -y curl && \
 curl -o helm.tgz https://storage.googleapis.com/kubernetes-helm/helm-v2.11.0-linux-amd64.tar.gz && \
 tar -xzf helm.tgz && \
 mv linux-amd64/helm /usr/local/bin && \
 rm helm.tgz
RUN helm init --client-only
```

In this case, the `helm` mixin will first run apt-get update and then install `curl` in the Docker image. Next, it will use curl to fetch Helm 2.11.0, extract it into the Docker filesystem and finally run `helm init --client-only` to do necessary setup inside the image.

### Run Time

The [CNAB specification](https://github.com/cnabio/cnab-spec/blob/master/103-bundle-runtime.md) specifies three actions that an invocation image should support:

* install
* upgrade
* uninstall

Porter in turn, expects that a mixin should provide a command that corresponds to each of these actions. If the corresponding action is not relevant, the mixin should still provide a command for the action and return no error. Here is `helm` mixin again for reference:

```console
$ ./bin/mixins/helm/helm
A helm mixin for porter 👩🏽‍✈️

Usage:
  helm [command]

Available Commands:
  build       Generate Dockerfile lines for the bundle invocation image
  help        Help about any command
  install     Execute the install functionality of this mixin
  uninstall   Execute the uninstall functionality of this mixin
  version     Print the mixin version

Flags:
      --debug   Enable debug logging
  -h, --help    help for helm

Use "helm [command] --help" for more information about a command.
```

Porter will pass the entire step, in YAML form, to the mixin. Porter expects the step YAML to have a `description` field and an array of optional `outputs`, and allows each mixin to process the remaining structure of the YAML as needed. For example, the `helm` mixin expects to be passed a YAML document like this:

```yaml
helm3:
  description: "Install MySQL"
  name: porter-ci-mysql
  chart: bitnami/mysql
  version: 6.14.2
  replace: true
  set:
    db.name: ${ bundle.parameters.database-name }
    db.user: ${ bundle.parameters.mysql-user }
  outputs:
  - name: mysql-root-password
    secret: porter-ci-mysql
    key: mysql-root-password
  - name: mysql-password
    secret: porter-ci-mysql
    key: mysql-password
```

In this case, the `description` and `outputs` elements of this YAML are defined by Porter, but the `helm` block is defined by the mixin. The `outputs` section itself is also largely defined by the mixin. Porter will pass the entire block to the mixin, with the expectation that the mixin will report errors if the YAML is incorrect. The next section discusses how mixins can define their inputs and outputs.

### Input and Output

Mixins can expose JSON schema to describe the YAML format that they expect to accept as input, including output format and any parameters that are required. Mixins should expose this JSON schema through a **TBD** command.

## Example Mixins

If you'd like to build a mixin and would like to refer to some existing implementations, please see:

* [helm-mixin](https://github.com/getporter/helm-mixin) - The Porter Helm Mixin
* [az-mixin](https://github.com/getporter/az-mixin) - The Porter Azure (az cli) Mixin