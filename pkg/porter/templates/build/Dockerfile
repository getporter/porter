FROM debian:stretch

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates

COPY .cnab /cnab
COPY . $BUNDLE_DIR
RUN rm -fr $BUNDLE_DIR/.cnab
