#!/usr/bin/env bash
set -euo pipefail

trap 'make -f Makefile.kind delete-kind-cluster' EXIT
make -f Makefile.kind install-kind create-kind-cluster
make start-local-docker-registry test-integration stop-local-docker-registry