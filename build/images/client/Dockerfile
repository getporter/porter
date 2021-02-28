FROM alpine:3

ARG PERMALINK

RUN apk add curl --no-cache
RUN curl --http1.1 -v -H "X-Azure-DebugInfo: 1" -A "curl build-porter-client porter_trace_$(date +%s)" https://cdn.porter.sh/${PERMALINK}/install-linux.sh | sh && \
    ln -s /root/.porter/porter /usr/local/bin/porter

ENTRYPOINT ["/root/.porter/porter"]
