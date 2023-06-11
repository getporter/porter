FROM alpine:3 as builder
ARG TARGETARCH
WORKDIR /app/.porter

RUN mkdir runtimes && \
  mkdir -p mixins/exec/runtimes

# Install porter-grpc-server, agent, and the exec mixin, everything else
COPY bin/dev/porter-grpc-server-linux-$TARGETARCH porter
COPY bin/dev/agent-linux-$TARGETARCH agent
COPY bin/mixins/exec/dev/exec-linux-$TARGETARCH mixins/exec/exec
RUN ln -s /app/.porter/porter runtimes/porter-runtime && \
  ln -s /app/.porter/mixins/exec/exec mixins/exec/runtimes/exec-runtime

# Copy the porter installation into a distroless container
# Explicitly not using the nonroot tag because we don't want the user to exist so it is placed in the root group
# This allows us to run with a random UID, and access a mounted docker socket (which is only accessible via the root group)
FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=builder --chown=65532:0 --chmod=770 /app/.porter /app/.porter
ENV PATH "$PATH:/app/.porter"
# This is where files that need to be copied into /app/.porter/ should be mounted
VOLUME /porter-config
ENV PORTER_HOME /app/.porter

# Run as a nonroot user
USER 65532
ENTRYPOINT ["/app/.porter/agent"]