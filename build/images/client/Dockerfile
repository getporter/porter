FROM alpine:3

RUN mkdir -p /root/.porter/runtimes && \
    mkdir -p /root/.porter/mixins/exec/runtimes

COPY bin/dev/porter-linux-amd64 /root/.porter/porter
COPY bin/mixins/exec/dev/exec-linux-amd64 /root/.porter/mixins/exec/exec
RUN ln -s /root/.porter/porter /root/.porter/runtimes/porter-runtime && \
    ln -s /root/.porter/mixins/exec/exec /root/.porter/mixins/exec/runtimes/exec-runtime && \
    ln -s /root/.porter/porter /usr/local/bin/porter

RUN porter mixin install kubernetes && \
    porter mixin install helm && \
    porter mixin install arm && \
    porter mixin install terraform && \
    porter mixin install az && \
    porter mixin install aws && \
    porter mixin install gcloud && \
    porter plugin install azure && \ 
    porter plugin install kubernetes

ENTRYPOINT ["porter"]
