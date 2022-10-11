#!/usr/bin/env bash

set -xeuo pipefail

export PATH=$PATH:~/.porter

PORTER_VERSION=canary ./scripts/install/install-mac.sh
porter version
# do not call list, the macos agent doesn't have docker installed

PORTER_VERSION=v0.23.0-beta.1 ./scripts/install/install-mac.sh
porter version | grep v0.23.0-beta.1

PORTER_VERSION=latest ./scripts/install/install-mac.sh
