FROM alpine:3

RUN mkdir -p /root/.porter/runtimes && \
    mkdir -p /root/.porter/mixins/exec/runtimes

COPY bin/dev/porter-linux-amd64 /root/.porter/porter
COPY bin/mixins/exec/dev/exec-linux-amd64 /root/.porter/mixins/exec/exec
RUN ln -s /root/.porter/porter /root/.porter/runtimes/porter-runtime && \
    ln -s /root/.porter/mixins/exec/exec /root/.porter/mixins/exec/runtimes/exec-runtime && \
    ln -s porter /usr/local/bin/porter

ENTRYPOINT ["porter"]
