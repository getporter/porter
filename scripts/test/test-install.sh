#!/usr/bin/env bash

set -xeuo pipefail

PORTER_VERSION=canary ./scripts/install/install-linux.sh
PORTER_VERSION=latest ./scripts/install/install-linux.sh