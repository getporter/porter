#!/usr/bin/env bash
set -euo pipefail

trap 'make -f Makefile.kind delete-kind-cluster' EXIT
make -f Makefile.kind install-kind create-kind-cluster
make test-integration