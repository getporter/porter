FROM quay.io/deis/lightweight-docker-go:v0.2.0
FROM debian:stretch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
