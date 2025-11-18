# syntax=docker/dockerfile-upstream:1.4.0
FROM --platform=linux/amd64 debian:stable-slim

# PORTER_INIT

RUN rm -f /etc/apt/apt.conf.d/docker-clean; echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache
RUN --mount=type=cache,target=/var/cache/apt --mount=type=cache,target=/var/lib/apt \
    apt-get update && apt-get install -y ca-certificates

# PORTER_MIXINS

# Copy user files from the bundle source directory (excludes .cnab and porter.yaml via .dockerignore)
COPY --from=userfiles --link . ${BUNDLE_DIR}/
