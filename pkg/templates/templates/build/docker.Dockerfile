FROM debian:stretch-slim

# PORTER_INIT

RUN apt-get update && apt-get install -y ca-certificates

# PORTER_MIXINS

COPY . ${BUNDLE_DIR}
