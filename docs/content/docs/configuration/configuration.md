---
title: Configuration
description: Controlling Porter with its config file, environment variables and flags
weight: 4
aliases:
  - /configuration
  - /configuration/
---

Porter has a hierarchical configuration system that loads configuration values in the following precedence order:

- Flags (highest)
- Environment Variables
- Config File (lowest)

You may set a default value for a configuration value in the config file, override it with an environment variable, and then override both for a particular command with a flag.

- [Flags](#flags)
- [Environment Variables](#environment-variables)
- [Config File](#config-file)
- [Experimental Feature Flags](#experimental-feature-flags)
  - [Build Drivers](#build-drivers)
  - [Structured Logs](#structured-logs)
  - [Dependencies v2](#dependencies-v2)
  - [Full control Dockerfile](#full-control-dockerfile)
- [Common Configuration Settings](#common-configuration-settings)
  - [Set Current Namespace](#namespace)
  - [Output Formatting](#output)
- [Allow Docker Host Access](#allow-docker-host-access)

## Flags

Nearly all of Porter's configuration, except global configuration such as secret accounts, storage accounts, or telemetry, are configurable by flags.
Use the `porter help` command to view available flags.

## Environment Variables

Flags have corresponding environment variables that you can use so that you don't need to manually set the flag every time.
The flag will default to the value of the environment variable, when defined.
Global configuration settings can also be specified with an environment variable.
For example, the experimental config file setting maps to PORTER_EXPERIMENTAL, and accepts a comma-separated list of values.

\--flag maps to the environment variable of PORTER_FLAG.
Dashes in the flag name are represented as underscores in the environment variable name.
So \--another-flag maps to the environment variable PORTER_ANOTHER_FLAG

For example, you can set PORTER_OUTPUT=json and then all subsequent porter commands will act as though the \--output=json flag was passed.

## Config File

Porter's configuration file is located in the PORTER_HOME directory, by default ~/.porter/.
The file name should be config.FILE_EXTENSION, where the file extension can be json, toml, yaml, or hcl.
For example, If you defined the configuration in YAML, the file is named config.yaml.

Do not embed sensitive data in the configuration file.
Instead, use templates to inject environment variables or secrets in the configuration file.
Environment variables are specified with ${env.NAME}, where name is case-sensitive.
Secrets are specified with ${secret.KEY} and case sensitivity depends upon the secrets plugin used.

### Multiple environments

Porter supports defining multiple named environments (contexts) in a single
config file, similar to how `kubectl` handles multiple clusters.
See [Multiple Configuration Environments](/docs/configuration/multi-context/) for details.

### Flat config file (legacy)

The flat format places all settings at the top level of the config file.
This format is still supported but is considered legacy.
Use `porter config migrate` to convert it to the multi-context format.

Below is a full example in YAML:

```yaml
# ~/.porter/config.yaml

# Set the default namespace
namespace: "dev"

# Threshold for printing messages to the console
# Allowed values are: debug, info, warn, error.
# Does not affect what is written to the log file or traces.
verbosity: "debug"

# Default command output to JSON
output: "json"

# Allow all bundles access to the Docker Host
allow-docker-host-access: true

# Enable experimental features
experimental:
  - "flagA"
  - "flagB"

# Use Docker buildkit to build the bundle
build-driver: "buildkit"

# Do not automatically build a bundle from source
# before running the requested command when Porter detects that it is out-of-date.
# Porter detects changes to porter.yaml, mixins, Porter version, and all files in the bundle directory
# (including scripts, templates, and custom Dockerfiles).
# Example: Normally running porter explain in a bundle directory should trigger an automatic porter build
# after you have edited porter.yaml or any bundle files, and disabling autobuild would have porter explain
# use the cached build (which could be stale).
autobuild-disabled: true

# Overwrite the existing published bundle when publishing or copying a bundle.
# By default, Porter detects when a push would overwrite an existing artifact and requires --force to proceed.
force-overwrite: false

# Use the storage configuration named devdb
default-storage: "devdb"

# When default-storage is not set, use the mongodb-docker plugin.
# This mode does not support additional configuration for the plugin.
# If the plugin requires configuration, use default-storage and define
# the configuration in the storage section.
default-storage-plugin: "mongodb-docker"

# Use the secrets configuration named mysecrets
default-secrets: "mysecrets"

# When default-secrets is not set, use the kubernetes.secret plugin.
# This mode does not support additional configuration for the plugin.
# If the plugin requires configuration, use default-secrets and define
# the configuration in the secrets section.
default-secrets-plugin: "kubernetes.secret"

# Use the signer configuration name mysigner.
# If not specified, bundles and bundle images cannot be signed.
default-signer: "mysigner"

# Defines storage accounts
storage:
  # The storage account name
  - name: "devdb"

    # The plugin used to access the storage account
    plugin: "mongodb"

    # Additional configuration for storage account
    # These values vary depending on the plugin used
    config:
      # The mongodb connection string
      url: "${secret.porter-db-connection-string}"

      # Timeout for database queries
      timeout: 300

# Define secret store accounts
secrets:
  # The secret store name
  - name: "mysecrets"

    # The plugin used to access the secret store account
    plugin: "azure.keyvault"

    # Additional configuration for secret store account
    # These values vary depending on the plugin used
    config:
      # The name of the secret vault
      vault: "topsecret"

      # The subscription where the vault is defined
      subscription-id: "${env.AZURE_SUBSCRIPTION_ID}"

# Define signers
signer:
  # The signer name
  - name: "mysigner"
    
    # The plugin used to sign bundles
    plugin: "cosign"
    
    # Additional configuration for the signer
    # These values vary depending on the plugin used
    config:
      # Path to the public key
      publickey: /home/porter/cosign.pub

      # Path to the public key
      privatekey: /home/porter/cosign.key

# Log command output to a file in PORTER_HOME/logs/
logs:
  # Log command output to a file
  log-to-file: true

  # When structured is true, the logs printed to the console
  # include a timestamp and log level
  structured: false

  # Sets the log level for what is written to the file
  # Allowed values: debug, info, warn, error
  level: "info"

# Send trace and log data to an Open Telemetry collector
telemetry:
  # Enable trace collection
  enabled: true

  # Send telemetry via the grpc protocol
  # Allowed values: http/protobuf, grpc
  protocol: "grpc"

  # The Open Telemetry collector endpoint
  endpoint: "127.0.0.1:4318"

  # Specify if the collector endpoint is secured with TLS
  insecure: true

  # Specify a certificate to connect to the collector endpoint
  certificate: "/home/me/some-cert.pem"

  # The compression type used when communicating with the collector endpoint
  compression: "gzip"

  # The timeout enforced when communicating with the collector endpoint
  timeout: "3s"

  # The timeout enforced when establishing a connection with the collector endpoint
  start-timeout: "100ms"

  # Used for testing that porter is emitting spans without setting up an open telemetry collector
  redirect-to-file: false

  # Additional headers to send to the open telemetry collector
  headers:
    environment: "dev"
    owner: "myusername"
```

## Experimental Feature Flags

Porter sometimes uses feature flags to release new functionality for users to
evaluate, without affecting the stability of Porter. You can enable an experimental
feature by:

- Using the experimental global flag `--experimental flagA,flagB`.
  The value is a comma-separated list of strings.
- Setting the PORTER_EXPERIMENTAL environment variable like so `PORTER_EXPERIMENTAL=flagA,flagB`.
  The value is a comma-separated list of strings.
- Setting the experimental field in the configuration file like so `experimental: ["flagA","flagB"]`.
  The value is an array of strings.

### Build Drivers

The **build-drivers** experimental feature flag is no longer active.
Build drivers are enabled by default and the only available driver is buildkit.

The docker driver uses the local Docker host to build a bundle image, and run it in a container.
To use a remote Docker host, set the following environment variables:

- DOCKER_HOST (required)
- DOCKER_TLS_VERIFY (optional)
- DOCKER_CERTS_PATH (optional)

### Structured Logs

The **structured-logs** experimental feature flag is no longer active.
Use the trace and logs configuration sections below to configure how logs and telemetry should be collected.

#### Logs

Porter can be configured to [write a logfile for each command](/docs/administration/collect-diag-porter/#logs).

The following log settings are available:

| Setting          | Environment Variable    | Description                                                                                                                                                                                                    |
| ---------------- | ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| logs.log-to-file | PORTER_LOGS_LOG_TO_FILE | Specifies if a logfile should be written for each command.                                                                                                                                                     |
| logs.structured  | PORTER_LOGS_STRUCTURED  | Specifies if the logs printed to the console should include a timestamp and log level                                                                                                                          |
| logs.level       | PORTER_LOGS_LEVEL       | Filters the logs to the specified level and higher. The log level controls the logs written to file when porter is run. Allowed values are: debug, info, warn, error. |

#### Telemetry

Porter supports the OpenTelemetry specification for exporting trace data.

Porter automatically uses the standard [OpenTelemetry environment variables][otel] to configure the trace exporter.

| Setting               | Environment Variable                                            | Description                                                                                                      |
| --------------------- | --------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| telemetry.enabled     | PORTER_TELEMETRY_ENABLED                                        | Enables telemetry collection. Defaults to false.                                                                 |
| telemetry.protocol    | OTEL_EXPORTER_OTLP_PROTOCOL<br/>PORTER_TELEMETRY_PROTOCOL       | The protocol used to connect with the telemetry server. Either grpc or http/protobuf. Defaults to http/protobuf. |
| telemetry.endpoint    | OTEL_EXPORTER_OTLP_ENDPOINT<br/>PORTER_TELEMETRY_ENDPOINT       | The endpoint where traces should be sent. Defaults 127.0.0.1:4317.                                               |
| telemetry.insecure    | OTEL_EXPORTER_OTLP_INSECURE<br/>PORTER_TELEMETRY_INSECURE       | If true, TLS is not used, which is useful for local development and self-signed certificates.                    |
| telemetry.certificate | OTEL_EXPORTER_OTLP_CERTIFICATE<br/>PORTER_TELEMETRY_CERTIFICATE | Path to the PEM formatted certificate to use with the endpoint.                                                  |
| telemetry.compression | OTEL_EXPORTER_OTLP_COMPRESSION<br/>PORTER_TELEMETRY_COMPRESSION | Supported values are: gzip. Defaults to no compression.                                                          |
| telemetry.timeout     | OTEL_EXPORTER_OTLP_TIMEOUT<br/>PORTER_TELEMETRY_TIMEOUT         | A timeout to use with the telemetry server, in Go duration format. For example, 30s or 1m.                       |
| telemetry.headers     | OTEL_EXPORTER_OTLP_HEADERS<br/>PORTER_TELEMETRY_HEADERS         | A map of key/value pairs that should be sent as headers to the telemetry server.                                 |

Below is a sample Porter configuration file that demonstrates how to set each of the telemetry settings:

```yaml
telemetry:
  enabled: true
  protocol: "grpc"
  endpoint: "127.0.0.1:4318"
  insecure: true
  certificate: "/home/me/some-cert.pem"
  compression: "gzip"
  timeout: "3s"
  start-timeout: "100ms"

  headers:
    environment: "dev"
    owner: "me"
```

[otel]: https://github.com/open-telemetry/opentelemetry-specification/blob/v1.8.0/specification/protocol/exporter.md

### Dependencies v2

The `dependencies-v2` experimental flag is not yet implemented.
When it is completed, it is used to activate the features from [PEP003 - Advanced Dependencies](https://github.com/getporter/proposals/blob/main/pep/003-advanced-dependencies.md).

### Full control Dockerfile

The `full-control-dockerfile` experimental flag disables all Dockerfile generation when building bundles.
When enabled Porter will use the file referenced by `dockerfile` in the Porter manifest when building the bundle image *without modifying* it in any way.
Ie. Porter will not process `# PORTER_x` placeholders, nor inject any user configuration and `CMD` statements.
It is up to the bundle author to ensure that the contents of the Dockerfile contains the necessary tools for any mixins to function and a layout that can be executed as a Porter bundle.

## Common Configuration Settings

Some configuration settings are applicable to many of Porter's commands and to save time you may want to set these values in the configuration file or with environment variables.

### Namespace

\--namespace specifies the current namespace.
It is set with the PORTER_NAMESPACE environment variable.

```yaml
namespace: "dev"
```

### Output

\--output controls the format of the command output printed by porter.
It is set with the PORTER_OUTPUT environment variable.
Each command supports a different set of allowed outputs though usually there is some combination of: plaintext, json, and yaml.

```yaml
output: "json"
```

### Allow Docker Host Access

\--allow-docker-host-access controls whether the local Docker daemon or host should be made available to executing bundles.
It is set with the PORTER_ALLOW_DOCKER_HOST_ACCESS environment variable.

This flag is available for the following commands: install, upgrade, invoke, and uninstall.
When this value is set to true, bundles are executed in a privileged container with the docker socket mounted.
This allows you to use Docker from within your bundle, such as `docker push`, `docker-compose`, or docker-in-docker.

🚨 **There are security implications to enabling access!
You should trust any bundles that you execute with this setting enabled as it gives them elevated access to the host machine.**

⚠️️ This configuration setting is only available when you are in an environment that provides access to the local docker daemon.
Therefore, it does not work with the Azure Cloud Shell driver.

### Schema Check

The schema-check configuration file setting controls Porter's behavior when the schemaVersion of a resource does not match [Porter's supported version](/reference/file-formats/).
By default, Porter requires that a resource's schemaVersion field matches Porter's allowed version(s).
In some cases, such as when migrating to a new version of Porter, it may be helpful to use a less strict version comparison.
Allowed values are:

- exact - Default behavior. Require that the schemaVersion on the resource exactly match Porter's supported version(s).
  If it doesn't match, the command will fail.
- minor - Require that the MAJOR.MINOR portion of the schemaVersion on the resource match Porter's supported version.
  For example, a bundle with a schemaVersion of 1.2.3 would work even though the supported version is 1.2.5.
- major - Require that the MAJOR portion of the schemaVersion on the resource match Porter's supported version.
  For example, a bundle with a schemaVersion of 1.2.3 would work even though the supported version is 1.3.0.
- none - Only print a warning when the schemaVersion does not exactly match the supported version.

Porter can only guarantee correct parsing of the file when the schemaVersion exactly matches.
Depending on what has changed between schema versions, you can make a judgement call on if those changes are relevant to your situation.
