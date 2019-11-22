FROM alpine:3

ARG PERMALINK

RUN apk add curl --no-cache
RUN sh -c 'curl https://cdn.deislabs.io/porter/${PERMALINK}/install-linux.sh | sh' && \
    ln -s /root/.porter/porter /usr/local/bin/porter

ENTRYPOINT ["/root/.porter/porter"]
