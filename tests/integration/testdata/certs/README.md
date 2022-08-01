# Test Certificates

This directory contains self-signed certificates for our tests.
We use them to run a Docker registry that is secured with TLS.

When the certificate expires, run the following command from inside this directory to make a new one.

```bash
openssl req -new -newkey rsa:4096 -days 365 -nodes \
  -x509 -keyout registry_auth.key -out registry_auth.crt \
  -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=www.example.com"
```
