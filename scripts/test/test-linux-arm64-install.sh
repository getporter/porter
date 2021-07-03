#!/usr/bin/env bash

set -xeuo pipefail

export PATH=$PATH:~/.porter

PORTER_PERMALINK=canary ./scripts/install/install-linux-arm64.sh
porter list

PORTER_PERMALINK=v0.23.0-beta.1 ./scripts/install/install-linux-arm64.sh
porter version | grep v0.23.0-beta.1

PORTER_PERMALINK=latest ./scripts/install/install-linux-arm64.sh
