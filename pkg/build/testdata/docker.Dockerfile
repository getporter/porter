FROM debian:stretch-slim

ARG BUNDLE_DIR

RUN groupadd nonroot -o -g 65532 &&\
    useradd nonroot -m -u 65532 -g 65532 -o

RUN apt-get update && apt-get install -y ca-certificates


COPY . $BUNDLE_DIR
RUN rm $BUNDLE_DIR/porter.yaml
RUN rm -fr $BUNDLE_DIR/.cnab
COPY .cnab /cnab
RUN chown -R nonroot.nonroot /cnab
USER 65532:65532
WORKDIR $BUNDLE_DIR
CMD ["/cnab/app/run"]