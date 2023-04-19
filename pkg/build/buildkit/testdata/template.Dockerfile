# syntax=docker/dockerfile-upstream:1.4.0
FROM --platform=linux/amd64 debian:stretch-slim

# PORTER_INIT

RUN rm -f /etc/apt/apt.conf.d/docker-clean; echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache
RUN --mount=type=cache,target=/var/cache/apt --mount=type=cache,target=/var/lib/apt \
    apt-get update && apt-get install -y ca-certificates

# PORTER_MIXINS

# Use the BUNDLE_DIR build argument to copy files into the bundle's working directory
COPY --link . ${BUNDLE_DIR}
