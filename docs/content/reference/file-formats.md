---
title: File Formats
description: Defines the format of files used by Porter
---

* [Supported Versions](#supported-versions)
* [Manifest](#manifest)
* [Credential Sets](#credential-set)
* [Parameter Sets](#parameter-set)
* [Installation](#installation)
* [Porter Operator File Formats](/operator/file-formats/)

## Supported Versions

Below are schema versions for each of the file formats, and the corresponding Porter version that supports it.

| Schema Type   | Schema Version | Porter Version   |
|---------------|----------------|------------------|
| Bundle        | (none)         | v0.38.*          |
| Bundle        | 1.0.0-alpha.1  | v1.0.0-alpha.14+ |
| CredentialSet | (none)         | v0.38.*          |
| CredentialSet | 1.0.1          | v1.0.0-alpha.1+  |
| ParameterSet  | (none)         | v0.38.*          |
| ParameterSet  | 1.0.1          | v1.0.0-alpha.1+  |
| Installation  | 1.0.0          | v1.0.0-alpha.1+  |

## Manifest

The manifest is the porter.yaml file used to build a bundle.
You can use this [json schema][manifest-schema] to validate a Porter manifest.

```yaml
schemaType: Bundle
schemaVersion: 1.0.0-alpha.1
name: myapp
version: 1.0.0
description: Install my great application
registry: localhost:5000
reference: localhost:5000/myapp:v1.0.0
dockerfile: template.Dockerfile

maintainers:
  - name: Qi Lu
  - email: sal@example.com
  - name: Frank
    url: https://example.com/frank

custom:
  app:
    version: 1.2.3
    commit: abc123

required:
  - docker:
      privileged: true

mixins:
  - exec
  - helm3:
      clientVersion: 3.1.2

credentials:
  - name: kubeconfig
    description: A kubeconfig with cluster admin role
    path: /root/.kube/config
  - name: token
    env: GITHUB_TOKEN
    applyTo:
      - release

parameters:
  - name: log-level
    description: Log level for MyApp
    type: integer
    env: MYAPP_LOG_LEVEL
    applyTo:
      - install
      - upgrade
  - name: connstr
    description: MyApp database connnection string
    type: string
    env: MYAPP_CONNECTION_STRING
    sensitive: true
    source:
      dependency: mysql
      output: admin-connstr
    applyTo:
      - install
      - status
  - name: release-name
    type: string
    env: RELEASE
    default: myapp

state:
  - name: tfstate
    description: Store terraform state with the bundle instead of a remote backend
    path: terraform/terraform.tfstate

dependencies:
  - name: mysql
    reference: getporter/mysql:v0.1.1
    parameters:
      database: myapp

outputs:
  - name: app-token
    description: Access token for MyApp
    type: file
    path: /cnab/app/myapp_token
    sensitive: true
    applyTo:
      - install
  - name: ip-address
    type: string

images:
  myapp:
    repository: example/myapp
    digest: sha256:568461508c8d220742add8abd226b33534d4269868df4b3178fae1cba3818a6e

install:
  - helm3:
      description: "Install MyApp"
      name: "{{ bundle.parameters.release }}"
      chart: ./charts/myapp
      replace: true
      set:
        image.repository: "{{ bundle.images.myapp.repository }}"
        image.digest: "{{ bundle.images.myapp.digest }}"

upgrade:
  - helm3:
      name: "{{ bundle.parameters.release }}"
      chart: ./charts/myapp
      set:
        image.repository: "{{ bundle.images.myapp.repository }}"
        image.digest: "{{ bundle.images.myapp.digest }}"

uninstall:
  - helm3:
      purge: true
      releases:
        - "{{ bundle.parameters.release }}"  

poke:
  - exec:
      command: ./poke-myapp.sh

customActions:
  status:
    description: See what's up in there
    modifies: false
    stateless: true

status:
  - exec:
      command: ./status.sh
```

| Field                            | Required | Description                                                                                                                                                                            |
|----------------------------------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| schemaType                       | false    | The type of document. This isn't used by Porter but is included when Porter outputs the file, so that editors can determine the resource type.                                         |
| schemaVersion                    | true     | The version of the Bundle schema used in this file.                                                                                                                                    |
| name                             | true     | The name of the bundle.                                                                                                                                                                |
| version                          | true     | The version of the bundle, must adhere to [semver v2].<br/>The bundle tag defaults to the version with a v prefix, e.g. mybundle:v1.0.0. Use --tag or the reference field to override. |
| description                      | false    | A description of the bundle.                                                                                                                                                           |
| registry                         | true*    | The OCI registry to use when the bundle is published.<br/> /*Either registry or reference must be specified.                                                                           |
| reference                        | true*    | The full reference to use when the bundle is published to an OCI registry.<br/> /*Either registry or reference must be specified.                                                      |
| dockerfile                       | false    | The relative path to a Dockerfile to use as a template during porter build.                                                                                                            |
| maintainers                      | false    | A list of maintainers for the bundle.                                                                                                                                                  |
| maintainers.name                 | false    | Name of the maintainer.                                                                                                                                                                |
| maintainers.email                | false    | Email of the maintainer.                                                                                                                                                               |
| maintainers.url                  | false    | Url of the maintainer.                                                                                                                                                                 |
| custom                           | false    | An object of custom bundle metadata.                                                                                                                                                   |
| required                         | false    | A list of additional features required by the bundle.                                                                                                                                  |
| required.docker                  | false    | Indicate that the bundle requires a Docker host/daemon.                                                                                                                                |
| required.docker.privileged       | false    | Run the bundle in a privileged container.                                                                                                                                              |
| mixins                           | true     | A list of mixins used in the bundle. May include additional configuration settings depending on the mixin.                                                                             |
| credentials                      | false    | A list of credentials used by the bundle.                                                                                                                                              |
| credentials.name                 | true     | The name of the credential.                                                                                                                                                            |
| credentials.description          | false    | A description of the credential, how its used, and how to find it.                                                                                                                     |
| credentials.required             | false    | Indicates if the credential is required.                                                                                                                                               |
| credentials.applyTo              | false    | A list of actions that apply to this item. When none are specified, all actions apply.                                                                                                 |
| credentials.env                  | true     | The environment variable name, such as MY_VALUE, into which the credential value is stored. Either env or path must be specified.                                                      |
| credentials.path                 | true     | The path inside the bundle where the credential value is stored. Either env or path must be specified.                                                                                 |
| parameters                       | false    | A list of parameters used by the bundle.                                                                                                                                               |
| parameters.name                  | true     | The name of the parameter.                                                                                                                                                             |
| parameters.description           | false    | A description of the parameter and how it is used.                                                                                                                                     |
| parameters.default               | false    | The default value of the parameter. When no default is set, the parameter is required.                                                                                                 |
| parameters.applyTo               | false    | A list of actions that apply to this item. When none are specified, all actions apply.                                                                                                 |
| parameters.env                   | true     | The environment variable name, such as MY_VALUE, into which the parameter value is stored. Either env or path must be specified.                                                       |
| parameters.path                  | true     | The path inside the bundle where the parameter value is stored. Either env or path must be specified.                                                                                  |
| parameters.type                  | true     | The parameter type. Allowed values are: string, number, integer, boolean, object, or file.                                                                                             |
| parameters.sensitive             | false    | Indicates whether this parameter's value is sensitive and should not be logged.                                                                                                        |
| parameters.source                | false    | Indicates that the parameter should get its value from an external source.                                                                                                             |
| parameters.source.output         | true     | An output name. The parameter's value is set to output's last value.                                                                                                                   |
| parameters.source.dependency     | false    | The name of the dependency that generated the output. If not set, the output must be generated by the current bundle.                                                                  |
| state                            | false    | State variables that are generated by the bundle and injected on subsequent runs.                                                                                                      |
| state.name                       | true     | The name of this state variable                                                                                                                                                        |
| state.description                | false    | Description of how the variable is used by the bundle.                                                                                                                                 |
| state.mixin                      | false    | The name of the mixin that generates and manages this state variable.                                                                                                                  |
| state.path                       | true     | The path inside of the bundle where the state variable data is stored.                                                                                                                 |
| outputs                          | false    | A list of output values produced by the bundle.                                                                                                                                        |
| outputs.name                     | true     | The name of the output.                                                                                                                                                                |
| outputs.description              | false    | A description of the output and what it represents.                                                                                                                                    |
| outputs.default                  | false    | The default value of the output. When no default is set, the output is required.                                                                                                       |
| outputs.applyTo                  | false    | A list of actions that apply to this item. When none are specified, all actions apply.                                                                                                 |
| outputs.path                     | true     | The path inside the bundle where the output value is stored.                                                                                                                           |
| outputs.type                     | true     | The output type. Allowed values are: string, number, integer, boolean, object, or file.                                                                                                |
| outputs.sensitive                | false    | Indicates whether this output's value is sensitive and should not be logged.                                                                                                           |
| images                           | false    | A map of images referenced in the bundle. When the bundle is relocated, the referenced images are copied to the new bundle location.                                                   |
| images.KEY.description           | false    | A description of the image and how it is used.                                                                                                                                         |
| images.KEY.repository            | true     | The repository portion of the image reference, i.e. docker.io/library/nginx                                                                                                            |
| images.KEY.digest                | true     | The repository digest of the image, e.g. sha256:cafebabe...                                                                                                                            |
| dependencies                     | false    | Other bundles that this bundle relies upon.                                                                                                                                            |
| dependencies.requires            | true     | A list of bundles that should be executed with this bundle.                                                                                                                            |  
| dependencies.requires.name       | true     | A name for the dependency, used to refer to the dependency using the template syntax `bundle.dependencies.NAME`.                                                                       |
| dependencies.requires.reference  | true     | The full bundle reference for the dependency in the format REGISTRY/NAME:TAG.                                                                                                          |
| dependencies.requires.parameters | false    | A map of parameter names to their value.                                                                                                                                               |
| customActions                    | false    | A map of action names to a custom action definition.                                                                                                                                   |
| customActions.NAME.description   | false    | A description of the action.                                                                                                                                                           |
| customActions.NAME.modifies      | false    | Specifies if the action will modify resources in any way.                                                                                                                              |
| customActions.NAME.stateless     | false    | Specifies that the action could be run before the bundle is installed and does not require credentials.                                                                                |

[semver v2]: https://semver.org/spec/v2.0.0.html

## Credential Set

Credential sets can be defined in either json or yaml.
You can use this [json schema][cs-schema] to validate a credential set file.

```yaml
schemaType: CredentialSet
schemaVersion: 1.0.1
name: mycreds
namespace: staging
labels:
  team: redteam
  owner: xianglu
credentials:
  - name: token
    source:
      env: GITHUB_TOKEN
  - name: kubeconfig
    source:
      path: $HOME/.kube/config
  - name: connStr
    source:
      secret: my-connection-string
```

| Field              | Required | Description                                                                                                                                    |
|--------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------|
| schemaType         | false    | The type of document. This isn't used by Porter but is included when Porter outputs the file, so that editors can determine the resource type. |
| schemaVersion      | true     | The version of the Credential Set schema used in this file.                                                                                    |
| name               | true     | The name of the credential set.                                                                                                                |
| namespace          | false    | The namespace in which the credential set is defined. Defaults to the empty (global) namespace.                                                |
| labels             | false    | A set of key-value pairs associated with the credential set.                                                                                   |
| credentials        | true     | A list of credentials and instructions for Porter to resolve the credential value.                                                             |
| credentials.name   | true     | The name of the credential as defined in the bundle.                                                                                           |
| credentials.source | true     | Specifies how the credential should be resolved. Must have only one child property:<br/> secret, value, env, path, or command                  |

## Parameter Set

Parameter sets can be defined in either json or yaml.
You can use this [json schema][ps-schema] to validate a parameter set file.

```yaml
schemaType: ParameterSet
schemaVersion: 1.0.1
name: myparams
namespace: staging
labels:
  team: redteam
  owner: xianglu
parameters:
  - name: color
    source:
      value: red
  - name: log-level
    source:
      env: LOG_LEVEL
  - name: connStr
    source:
      secret: my-connection-string
```

| Field             | Required | Description                                                                                                                                    |
|-------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------|
| schemaType        | false    | The type of document. This isn't used by Porter but is included when Porter outputs the file, so that editors can determine the resource type. |
| schemaVersion     | true     | The version of the Parameter Set schema used in this file.                                                                                     |
| name              | true     | The name of the parameter set.                                                                                                                 |
| namespace         | false    | The namespace in which the parameter set is defined. Defaults to the empty (global) namespace.                                                 |
| labels            | false    | A set of key-value pairs associated with the parameter set.                                                                                    |
| parameters        | true     | A list of parameters and instructions for Porter to resolve the parameter value.                                                               |
| parameters.name   | true     | The name of the parameter as defined in the bundle.                                                                                            |
| parameters.source | true     | Specifies how the parameter should be resolved. Must have only one child property:<br/> secret, value, env, path, or command                   |

## Installation

Installations can be defined in either json or yaml.
You can use this [json schema][inst-schema] to validate an installation file.

Either the bundle digest, version, or tag must be specified.
When more than one is specified, Porter selects the most specific field available, preferring digest the most, then version, and then falling back to tag last.

```yaml
schemaType: Installation
schemaVersion: 1.0.0
name: myinstallation
namespace: staging
uninstalled: false
labels:
  team: marketing
  customer: bigbucks
bundle:
  repository: getporter/porter-hello
  # One of the following fields must be specified: digest, version, or tag
  digest: sha256:ace0eda3e3be35a979cec764a3321b4c7d0b9e4bb3094d20d3ff6782961a8d54
  version: 0.1.1
  tag: latest
parameterSets:
  - myparams
credentialSets:
  - mycreds
parameters:
  log-level: 11
```

| Field             | Required | Description                                                                                                                                    |
|-------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------|
| schemaType        | false    | The type of document. This isn't used by Porter but is included when Porter outputs the file, so that editors can determine the resource type. |
| schemaVersion     | true     | The version of the Installation schema used in this file.                                                                                      |
| name              | true     | The name of the installation.                                                                                                                  |
| namespace         | false    | The namespace in which the installation is defined. Defaults to the empty (global) namespace.                                                  |
| uninstalled       | false    | Specifies if the installation should be uninstalled. Defaults to false.                                                                        |
| labels            | false    | A set of key-value pairs associated with the installation.                                                                                     |
| bundle            | true     | A reference to where the bundle is published                                                                                                   |
| bundle.repository | true     | The repository where the bundle is published.                                                                                                  | 
| bundle.digest     | false*   | The bundle repository digest.                                                                                                                  |
| bundle.version    | false*   | The bundle version.                                                                                                                            |
| bundle.tag        | false*   | The bundle tag. This is useful when you do not use Porter's convention of defaulting the bundle tag to the bundle version.                     |
| parameterSets     | false    | A list of parameter set names.                                                                                                                 |
| credentialSets    | false    | A list of credential set names.                                                                                                                |
| parameters        | false    | Additional parameter values to use with the installation. Overrides any parameters defined in the associated parameter sets.                   |

\* The bundle section requires a repository and one of the following fields: digest, version, or tag.

[manifest-schema]: https://raw.githubusercontent.com/getporter/porter/release/v1/pkg/schema/manifest.schema.json
[cs-schema]: https://porter.sh/schema/v1/credential-set.schema.json
[ps-schema]: https://porter.sh/schema/v1/parameter-set.schema.json
[inst-schema]: https://porter.sh/schema/v1/installation.schema.json
