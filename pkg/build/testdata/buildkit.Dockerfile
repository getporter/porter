# syntax=docker/dockerfile-upstream:1.4.0
FROM --platform=linux/amd64 debian:stretch-slim

ARG BUNDLE_DIR
ARG BUNDLE_UID=65532
ARG BUNDLE_USER=nonroot
ARG BUNDLE_GID=0
RUN useradd ${BUNDLE_USER} -m -u ${BUNDLE_UID} -g ${BUNDLE_GID} -o

RUN rm -f /etc/apt/apt.conf.d/docker-clean; echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache
RUN --mount=type=cache,target=/var/cache/apt --mount=type=cache,target=/var/lib/apt \
    apt-get update && apt-get install -y ca-certificates


COPY --link . ${BUNDLE_DIR}
RUN rm ${BUNDLE_DIR}/porter.yaml
RUN rm -fr ${BUNDLE_DIR}/.cnab
COPY --link .cnab /cnab
RUN chgrp -R ${BUNDLE_GID} /cnab && chmod -R g=u /cnab
USER ${BUNDLE_UID}
WORKDIR ${BUNDLE_DIR}
CMD ["/cnab/app/run"]