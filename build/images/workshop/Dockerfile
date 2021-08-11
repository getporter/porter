FROM docker:dind

ENV HELM_VER v2.17.0

RUN mkdir -p /root/.porter/runtimes && \
    mkdir -p /root/.porter/mixins/exec/runtimes

RUN apk add bash \
            git \
            curl \
            bash-completion \
            jq \
            ca-certificates && \
    curl -o helm.tgz https://get.helm.sh/helm-${HELM_VER}-linux-amd64.tar.gz && \
    tar -xzf helm.tgz && \
    mv linux-amd64/helm /usr/local/bin && \
    rm helm.tgz && \
    helm init --client-only && \
    mkdir -p /workshop

COPY bin/dev/porter-linux-amd64 /root/.porter/porter
COPY bin/mixins/exec/dev/exec-linux-amd64 /root/.porter/mixins/exec/exec
RUN ln -s /root/.porter/porter /root/.porter/runtimes/porter-runtime && \
    ln -s /root/.porter/mixins/exec/exec /root/.porter/mixins/exec/runtimes/exec-runtime && \
    ln -s /root/.porter/porter /usr/local/bin/porter

WORKDIR /workshop
