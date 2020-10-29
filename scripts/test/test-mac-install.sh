#!/usr/bin/env bash

set -xeuo pipefail

export PATH=$PATH:~/.porter

PORTER_PERMALINK=canary ./scripts/install/install-mac.sh
porter list

PORTER_PERMALINK=v0.23.0-beta.1 ./scripts/install/install-mac.sh
porter version | grep v0.23.0-beta.1

PORTER_PERMALINK=latest ./scripts/install/install-mac.sh
