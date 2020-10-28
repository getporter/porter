---
title: "Example Bundle Format in OCI"
description: A sample CNAB bundle stored in an OCI registry
---

# What a CNAB bundle looks like in an OCI Registry

```
// layer 96d - this is the end manifest list that is digested and referenceable
{
    "schemaVersion": 2,
    "manifests": [
        {
            "mediaType": "application/vnd.oci.image.manifest.v1+json",
            "digest": "sha256:464e2efbee1cfa84d29b3305f0901c75dc70f2fa554cbcb7a342e21cf7d7f5e1",
            "size": 188,
            "annotations": {
                "io.cnab.manifest.type": "config"
            }
        },
        {
            "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
            "digest": "sha256:28ef97b8686a0b5399129e9b763d5b7e5ff03576aa5580d6f4182a49c5fe1913",
            "size": 2364,
            "annotations": {
                "io.cnab.manifest.type": "invocation"
            }
        }
    ],
    "annotations": {
        "io.cnab.runtime_version": "v1.0.0-WD",
        "io.docker.app.format": "cnab",
        "io.docker.type": "app",
        "org.opencontainers.artifactType": "application/vnd.cnab.manifest.v1",
        "org.opencontainers.image.authors": "[{\"name\":\"Matt Butcher\",\"email\":\"matt.butcher@microsoft.com\",\"url\":\"https://example.com\"}]",
        "org.opencontainers.image.description": "An example 'thin' helloworld Cloud-Native Application Bundle",
        "org.opencontainers.image.title": "helloworld-testdata",
        "org.opencontainers.image.version": "0.1.2"
    }
}
// layer 464e - this is the manifest that points to the parameters and credentials
{
    "schemaVersion": 2,
    "config": {
        "mediaType": "application/vnd.cnab.config.v1+json",
        "digest": "sha256:2224a999b6b9796820775b50299b6e4775c2fc978eac66713262175a3795bc6c",
        "size": 187
    },
    "layers": null
}
// layer 2224a - parameters and credentials
{
    "schema_version": "v1.0.0-WD",
    "parameters": {
        "backend_port": {
            "type": "integer",
            "destination": {
                "env": "BACKEND_PORT"
            }
        }
    },
    "credentials": {
        "hostkey": {
            "path": "/etc/hostkey.txt",
            "env": "HOST_KEY"
        }
    }
}
```