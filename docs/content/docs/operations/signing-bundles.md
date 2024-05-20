---
title: Signing Bundles
description: Signing of Porter bundles
weight: 7
---

Porter has built-in support for signing bundles and the associated invocation image using [Cosign] or [Notation].
Learn how to configure Porter to sign bundles.

- [Cosign](#cosign)
- [Notation](#notation)

## Cofiguration

### Cosign

#### Prerequisites

1. Cosign is installed and is available on the on the `PATH`.
2. A key-pair for signing is available.

Instructions on for the install Cosign can be found on the [Cosign Installation page](https://docs.sigstore.dev/system_config/installation/), and instructions on how to generate a key-pair can be found in the [Cosign Quickstart Guide](https://docs.sigstore.dev/signing/quickstart/#signing-with-a-generated-key).

🚧 Currently Porter does not support [Keyless Signing](https://docs.sigstore.dev/signing/overview/).

#### Configuration

Porter have to be configure to use [Cosign] to sign bundles and invocation images. All configuration is done through the [Porter config file](/docs/configuration/configuration/). To configure [Cosign] add the following to the configuration file.

```yaml
# ~/.porter/config.yaml

default-signer: "mysigner"

signer:
  - name: "mysigner"
    plugin: "cosign"
    config:
      publickey: <PATH_TO_PUBLIC_KEY>
      privatekey: <PATH_TO_PRIVATE_KEY>
      
      # Set the mode for fetching references from the registry. allowed: legacy, oci-1-1.
      # If set to oci-1-1, experimental must be set the true.
      # registrymode: legacy
      
      # Enable Cosign experimental features.
      # Required if regsitrymode is set to oci-1-1.
      # experimental: false
      
      # Allow signing of bundles in registries with expired or self-signed certificates.
      # insecureregistry: false
```

### Notation

#### Prerequisites

1. Notation is installed and is available on the on the `PATH`.
2. A signing key and certificate have been configured.
3. A trust policy for verification have been configured.

Instructions on for the install Notation can be found on the [Notation Installation page](https://notaryproject.dev/docs/user-guides/installation/cli/), and instructions on how to configure a signing key, certificate and trust policy can be found in the [Notation Quickstart Guide](https://notaryproject.dev/docs/quickstart-guides/quickstart-sign-image-artifact/).

#### Configuration

Porter have to be configure to use [Notation] to sign bundles and invocation images. All configuration is done through the [Porter config file](/docs/configuration/configuration/). To configure [Notation] add the following to the configuration file.

```yaml
# ~/.porter/config.yaml

default-signer: "mysigner"

signer:
  - name: "mysigner"
    plugin: "notation"
    config:
      key: <NAME_OF_KEY>

      # Allow signing of bundles HTTP registries
      # insecureregistry: false
```

## Sign bundle

To sign run [porter publish](/cli/porter_publish/) with the `--sign-bundle` flag.

## Verify bundle

A bundle can be verified before installation by adding the `--verify-bundle` flag to [porter install](/cli/porter_publish/).

[Cosign]: https://docs.sigstore.dev/signing/quickstart/
[Notation]: https://notaryproject.dev/docs/quickstart-guides/quickstart-sign-image-artifact/