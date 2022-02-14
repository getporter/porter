FROM debian:stretch-slim

ARG BUNDLE_DIR

RUN groupadd nonroot -o -g 65532 &&\
    useradd nonroot -m -u 65532 -g 65532 -o

RUN apt-get update && apt-get install -y ca-certificates

# PORTER_MIXINS

COPY . $BUNDLE_DIR
