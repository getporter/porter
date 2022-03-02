FROM debian:stretch-slim

ARG BUNDLE_DIR
ARG BUNDLE_UID=65532
ARG BUNDLE_USER=nonroot
ARG BUNDLE_GID=0
RUN useradd ${BUNDLE_USER} -m -u ${BUNDLE_UID} -g ${BUNDLE_GID} -o

RUN apt-get update && apt-get install -y ca-certificates


COPY . ${BUNDLE_DIR}
RUN rm ${BUNDLE_DIR}/porter.yaml
RUN rm -fr ${BUNDLE_DIR}/.cnab
COPY .cnab /cnab
RUN chgrp -R ${BUNDLE_GID} /cnab && chmod -R g=u /cnab
USER ${BUNDLE_UID}
WORKDIR ${BUNDLE_DIR}
CMD ["/cnab/app/run"]