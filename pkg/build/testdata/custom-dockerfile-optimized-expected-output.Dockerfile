# syntax=docker/dockerfile:1
FROM ubuntu:latest
# stuff
ARG BUNDLE_DIR
ARG BUNDLE_UID=65532
ARG BUNDLE_USER=nonroot
ARG BUNDLE_GID=0
RUN useradd ${BUNDLE_USER} -m -u ${BUNDLE_UID} -g ${BUNDLE_GID} -o
COPY mybin /cnab/app/

# exec mixin has no buildtime dependencies

COPY --from=porter-internal-userfiles --link . ${BUNDLE_DIR}/
COPY --link --chown=${BUNDLE_UID}:${BUNDLE_GID} --chmod=775 . /cnab
USER ${BUNDLE_UID}
WORKDIR ${BUNDLE_DIR}
CMD ["/cnab/app/run"]