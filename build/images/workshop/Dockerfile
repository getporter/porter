FROM docker:dind

ARG PERMALINK
ENV HELM_VER 2.17.0

RUN apk add bash \
            git \
            curl \
            bash-completion \
            jq \
            ca-certificates && \
    curl --http1.1 -v -H "X-Azure-DebugInfo: 1" -A "curl build-porter-workshop porter_trace_$(date +%s)" https://cdn.porter.sh/${PERMALINK}/install-linux.sh | bash && \
    ln -s /root/.porter/porter /usr/local/bin/porter && \
    curl -o helm.tgz https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VER}-linux-amd64.tar.gz && \
    tar -xzf helm.tgz && \
    mv linux-amd64/helm /usr/local/bin && \
    rm helm.tgz && \
    helm init --client-only && \
    mkdir -p /workshop

WORKDIR /workshop
