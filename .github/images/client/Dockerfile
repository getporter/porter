FROM alpine:3 as builder
WORKDIR /app/.porter

RUN mkdir runtimes && \
    mkdir -p mixins/exec/runtimes

# Only install porter and the exec mixin, everything else
# must be mounted into the container
COPY bin/dev/porter-linux-amd64 porter
COPY bin/mixins/exec/dev/exec-linux-amd64 mixins/exec/exec
RUN ln -s /app/.porter/porter runtimes/porter-runtime && \
    ln -s /app/.porter/mixins/exec/exec mixins/exec/runtimes/exec-runtime

# Copy the porter installation into a distroless container
# Explicitly not using the nonroot tag because we don't want the user to exist so it is placed in the root group
# This allows us to run with a random UID, and access a mounted docker socket (which is only accessible via the root group)
FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=builder --chown=65532:0 --chmod=770 /app/.porter /app/.porter
ENV PATH "$PATH:/app/.porter"

# Run as a nonroot user
USER 65532
ENTRYPOINT ["/app/.porter/porter"]
