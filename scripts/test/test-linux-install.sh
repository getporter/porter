#!/usr/bin/env bash

set -xeuo pipefail

export PATH=$PATH:~/.porter

PORTER_PERMALINK=canary ./scripts/install/install-linux.sh

PORTER_PERMALINK=v0.23.0-beta.1 ./scripts/install/install-linux.sh
porter version | grep v0.23.0-beta.1

PORTER_PERMALINK=latest ./scripts/install/install-linux.sh
