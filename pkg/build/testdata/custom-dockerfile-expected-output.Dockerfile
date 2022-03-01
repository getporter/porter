FROM ubuntu:latest
ARG BUNDLE_DIR
COPY mybin /cnab/app/

RUN rm $BUNDLE_DIR/porter.yaml
RUN rm -fr $BUNDLE_DIR/.cnab
COPY .cnab /cnab
RUN chgrp -R 0 /cnab && chmod -R g=u /cnab
USER 65532
WORKDIR $BUNDLE_DIR
CMD ["/cnab/app/run"]