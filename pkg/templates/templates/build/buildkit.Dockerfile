# syntax=docker/dockerfile:1.2
FROM debian:stretch-slim

ARG BUNDLE_DIR

RUN groupadd nonroot -o -g 65532 &&\
    useradd nonroot -m -u 65532 -g 65532 -o

RUN rm -f /etc/apt/apt.conf.d/docker-clean; echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache
RUN --mount=type=cache,target=/var/cache/apt --mount=type=cache,target=/var/lib/apt \
    apt-get update && apt-get install -y ca-certificates

# PORTER_MIXINS

COPY . $BUNDLE_DIR
