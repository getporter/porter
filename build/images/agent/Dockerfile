ARG PORTER_VERSION
ARG REGISTRY
FROM $REGISTRY/porter:$PORTER_VERSION

# This is where files that need to be copied into /app/.porter/ should be mounted
VOLUME /porter-config

ENV PORTER_HOME /app/.porter
COPY --chown=65532:0 --chmod=770 bin/dev/agent-linux-amd64 /app/.porter/agent

USER 65532
ENTRYPOINT ["/app/.porter/agent"]
