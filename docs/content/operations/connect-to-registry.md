---
title: Connect to a Registry
description: Configure Porter to authenticate and connect to a registry
---

Porter stores bundles in OCI (Docker) registries.
Learn how to configure Porter to authenticate and connect to your registry.

* [Authenticate to a Registry](#authenticate-to-a-registry)
* [Connect to an Insecure Registry](#connect-to-an-insecure-registry)
  * [Prerequisites](#prerequisites)
  * [Connect to an Unsecured Registry](#connect-to-an-unsecured-registry)
  * [Registry Secured with an Untrusted Certificate](#connect-to-a-registry-secured-with-an-untrusted-certificate)

## Authenticate to a Registry

Porter uses Docker's cached credentials to authenticate to a registry.
Before running a Porter command that requires authentication, first run `docker login REGISTRY` to authenticate.
For example, use `docker login` to authenticate to Docker Hub or `docker login ghcr.io` for GitHub Container Registry.

## Connect to an Insecure Registry

There are two situations where Porter considers a registry to be insecure: [the registry is not secured with TLS](#connect-to-an-unsecured-registry) or [the registry uses an untrusted TLS certificate](#connect-to-a-registry-secured-with-an-untrusted-certificate). 
The \--insecure-registry flag can be specified for any command that accepts a \--reference flag.

### Prerequisites

Your Docker engine must be [configured with any insecure registries](https://docs.docker.com/registry/insecure/).

### Connect to an Unsecured Registry

An unsecured registry communicates using HTTP instead of HTTPS.
Porter automatically uses HTTP to communicate with local registries running on the loopback IP address, such as localhost or 127.x.x.x.
No further configuration is required.

You can try this out by running a local Docker registry in a container:

```bash
# Run an unsecured Docker registry
docker run --name registry -d -p 0.0.0.0:5001:5000 registry:2

# Copy a bundle to it
porter copy --source ghcr.io/getporter/examples/porter-hello:v0.2.0 --destination localhost:5001/hello:v0.2.0

# Interact with the bundle
porter explain localhost:5001/hello:v0.2.0
```

If the registry is hosted on a non-loopback ip address or a domain name, the \--insecure-registry flag must be specified to allow connecting to the registry.

### Connect to a Registry Secured with an Untrusted Certificate

Sometimes you may want to connect to a registry that is secured with an untrusted TLS certificate.
For example, when you are running a test or development registry that uses a self-signed certificate or when using a custom root certificate that the system does not trust.
Use the the \--insecure-registry flag to allow connecting to the registry.

1. Create a self-signed certificate*

   **Bash**
   ```bash
   mkdir certs
   openssl req -new -newkey rsa:4096 -days 365 -nodes \
     -x509 -keyout certs/registry_auth.key -out certs/registry_auth.crt \
     -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=www.example.com"
   ```

   \* Apologies, we are still figuring out how to do the same thing on PowerShell,
   so please use WSL to accomplish this next task. If you know, please share so we can improve our documentation!
2. Run a Docker registry that uses your certificate generated in the previous step

   **Bash**
   ```bash
   docker run --name registry-with-tls -d \
     -v `pwd`/certs:/certs \
     -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/registry_auth.crt \
     -e REGISTRY_HTTP_TLS_KEY=/certs/registry_auth.key \
     -p 0.0.0.0:5002:5000 registry:2
   ```
   
   **PowerShell**
   ```powershell
   docker run --name registry-with-tls -d `
     -v ${pwd}/certs:/certs `
     -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/registry_auth.crt `
     -e REGISTRY_HTTP_TLS_KEY=/certs/registry_auth.key `
     -p 0.0.0.0:5002:5000 registry:2
   ```
3. Copy a bundle into your registry
   ```bash
   porter copy --insecure-registry --source ghcr.io/getporter/examples/porter-hello:v0.2.0 --destination localhost:5002/hello:v0.2.0
   ```
4. Interact with the bundle
   ```bash
   porter explain --insecure-registry localhost:5002/hello:v0.2.0
   ```

## Cleanup

Run the following commands to clean up resources created by the commands above:

```bash
# Remove the registry containers and their temporary volumes
docker rm -vf registry registry-with-tls
```

## Next Steps
* [Connect to Docker](/end-users/connect-docker/)
* [Configure Porter](/end-users/configuration/)
