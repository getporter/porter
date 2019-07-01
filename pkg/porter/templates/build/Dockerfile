FROM quay.io/deis/lightweight-docker-go:v0.2.0
FROM debian:stretch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ARG BUNDLE_DIR

COPY .cnab /cnab
COPY . $BUNDLE_DIR
RUN rm -fr $BUNDLE_DIR/.cnab
