# syntax=docker/dockerfile-upstream:1.4.0
FROM ubuntu:light
ARG BUNDLE_DIR
ARG BUNDLE_UID=65532
ARG BUNDLE_USER=nonroot
ARG BUNDLE_GID=0
RUN useradd ${BUNDLE_USER} -m -u ${BUNDLE_UID} -g ${BUNDLE_GID} -o
ARG BUNDLE_DIR
COPY mybin /cnab/app/
# exec mixin has no buildtime dependencies

COPY --link --chown=${BUNDLE_UID}:${BUNDLE_GID} --chmod=775 . /cnab
USER ${BUNDLE_UID}
WORKDIR ${BUNDLE_DIR}
CMD ["/cnab/app/run"]