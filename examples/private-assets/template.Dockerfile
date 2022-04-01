# syntax=docker/dockerfile-upstream:1.4.0
FROM debian:stretch-slim

# PORTER_INIT

RUN rm -f /etc/apt/apt.conf.d/docker-clean; echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache
RUN --mount=type=cache,target=/var/cache/apt --mount=type=cache,target=/var/lib/apt \
    apt-get update && apt-get install -y ca-certificates curl

# PORTER_MIXINS

# Use the BUNDLE_DIR build argument to copy files into the bundle's working directory
COPY --link . ${BUNDLE_DIR}

# Check the secret was passed to the build command
RUN --mount=type=secret,id=token /cnab/app/check-secrets.sh

# Use the injected secrets to build private assets into the bundle
RUN --mount=type=secret,id=token curl -O https://$(cat /run/secrets/token)@gist.githubusercontent.com/carolynvs/860a0d26de3af1468d290a075a91aac9/raw/c53223acd284830e8f541cf35eba94dde0ddf75d/secret
