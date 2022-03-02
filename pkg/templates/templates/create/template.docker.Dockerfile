# This is a template Dockerfile for the bundle's invocation image
# You can customize it to use different base images, install tools and copy configuration files.
#
# Porter will use it as a template and append lines to it for the mixins
# and to set the CMD appropriately for the CNAB specification.
#
# Add the following line to porter.yaml to instruct Porter to use this template
# dockerfile: template.Dockerfile

# You can control where the mixin's Dockerfile lines are inserted into this file by moving the "# PORTER_*" tokens
# another location in this file. If you remove a token, its content is appended to the end of the Dockerfile.
FROM debian:stretch-slim

# PORTER_INIT

RUN apt-get update && apt-get install -y ca-certificates

# PORTER_MIXINS

# Use the BUNDLE_DIR build argument to copy files into the bundle's working directory
COPY . ${BUNDLE_DIR}
