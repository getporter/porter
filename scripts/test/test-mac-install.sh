#!/bin/bash

set -euo pipefail

installer="$(cd "$(dirname "$0")/../install" && pwd)/install-mac.sh"
test_root=$(mktemp -d)
trap 'rm -rf "$test_root"' EXIT

expected_arch=$(uname -m)
if [[ "$expected_arch" == x86_64 && "$(sysctl -in sysctl.proc_translated 2>/dev/null || true)" == 1 ]]; then
    expected_arch=arm64
fi
versions=(v1.6.1 latest canary)
if [[ "$expected_arch" == x86_64 ]]; then
    versions+=(v0.23.0-beta.1)
fi

for version in "${versions[@]}"; do
    export PORTER_HOME
    PORTER_HOME=$(mktemp -d "$test_root/$version.XXXXXX")
    PORTER_VERSION="$version" /bin/bash "$installer"

    cli_version=$("$PORTER_HOME/porter" version)
    exec_version=$("$PORTER_HOME/mixins/exec/exec" version)
    echo "$cli_version"
    echo "$exec_version"
    if [[ "$version" == v* ]]; then
        echo "$cli_version" | grep -F "$version"
        echo "$exec_version" | grep -F "$version"
    fi

    file "$PORTER_HOME/porter" "$PORTER_HOME/mixins/exec/exec" "$PORTER_HOME/runtimes/porter-runtime"
    [[ "$(lipo -archs "$PORTER_HOME/porter")" == "$expected_arch" ]]
    [[ "$(lipo -archs "$PORTER_HOME/mixins/exec/exec")" == "$expected_arch" ]]
    file "$PORTER_HOME/runtimes/porter-runtime" | grep -E 'ELF 64-bit LSB.*x86-64'
done
