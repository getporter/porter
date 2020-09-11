# syntax=docker/dockerfile:experimental
# This is a multi-stage Dockerfile (build and run)
# Remember to target specific version in your base image
FROM alpine:3.12.0 AS build
# The Hugo version
ARG VERSION=0.53
# Get and install hugo
ADD https://github.com/gohugoio/hugo/releases/download/v${VERSION}/hugo_${VERSION}_Linux-64bit.tar.gz /hugo.tar.gz
RUN tar -zxvf hugo.tar.gz && mv hugo /usr/local/bin/hugo
RUN hugo version

# stage 2 serve hugo 
FROM alpine:3.12.0
# We add git to the build stage, because Hugo needs it
RUN apk add --no-cache git && mkdir -p /docs
# copy hugo binary from build
COPY --from=build /usr/local/bin/hugo /usr/local/bin/hugo
WORKDIR /docs
# Hadnle folder security 
RUN addgroup -Sg 1000 hugo && adduser -SG hugo -u 1000 -h /docs hugo
# Expose default hugo port
EXPOSE 1313

