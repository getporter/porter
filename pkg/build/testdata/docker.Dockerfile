FROM debian:stretch-slim

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates


COPY . $BUNDLE_DIR
RUN rm $BUNDLE_DIR/porter.yaml
RUN rm -fr $BUNDLE_DIR/.cnab
COPY .cnab /cnab
WORKDIR $BUNDLE_DIR
CMD ["/cnab/app/run"]