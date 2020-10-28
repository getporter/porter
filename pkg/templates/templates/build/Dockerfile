FROM debian:stretch

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates

# PORTER_MIXINS

COPY . $BUNDLE_DIR
