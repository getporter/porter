#!/usr/bin/env bash
set -euo pipefail

# Installs the porter CLI for a single user.
# PORTER_HOME:      Location where Porter is installed (defaults to ~/.porter).
# PORTER_MIRROR:    Base URL where Porter assets, such as binaries and atom feeds, are downloaded.
#                   This lets you setup an internal mirror.
# PORTER_VERSION:   The version of Porter assets to download.

export PORTER_HOME=${PORTER_HOME:-~/.porter}
export PORTER_MIRROR=${PORTER_MIRROR:-https://cdn.porter.sh}
PORTER_VERSION=${PORTER_VERSION:-latest}

arch=$(uname -m)
case "$arch" in
    x86_64)
        arch=amd64
        # A translated shell reports x86_64 on Apple Silicon. Intel Macs may
        # not have this sysctl key, so only an explicit 1 overrides amd64.
        if [[ "$(sysctl -in sysctl.proc_translated 2>/dev/null || true)" == 1 ]]; then
            arch=arm64
        fi
        ;;
    arm64) ;;
    *)
        echo "Unsupported macOS architecture: $arch" >&2
        exit 1
        ;;
esac

cli_temp=
runtime_temp=
cleanup() {
    local status=$?
    trap - EXIT
    if [[ -n "$cli_temp" ]]; then rm -f "$cli_temp" || true; fi
    if [[ -n "$runtime_temp" ]]; then rm -f "$runtime_temp" || true; fi
    exit "$status"
}
trap cleanup EXIT
trap 'status=$?; echo "Porter installation failed." >&2; exit "$status"' ERR

echo "Installing porter@$PORTER_VERSION to $PORTER_HOME from $PORTER_MIRROR"

mkdir -p "$PORTER_HOME/runtimes"

# Stage on each destination filesystem so each rename is atomic. Validate the
# CLI before replacing either binary; there is no rollback after replacement.
cli_temp=$(mktemp "$PORTER_HOME/.porter.XXXXXX")
runtime_temp=$(mktemp "$PORTER_HOME/runtimes/.porter-runtime.XXXXXX")
cli_url="$PORTER_MIRROR/download/$PORTER_VERSION/porter-darwin-$arch"
runtime_url="$PORTER_MIRROR/download/$PORTER_VERSION/porter-linux-amd64"

curl -fsSLo "$cli_temp" "$cli_url" || {
    status=$?
    echo "Failed to download Porter $PORTER_VERSION for darwin-$arch from $cli_url" >&2
    exit "$status"
}
curl -fsSLo "$runtime_temp" "$runtime_url" || {
    status=$?
    echo "Failed to download Porter $PORTER_VERSION runtime for linux-amd64 from $runtime_url" >&2
    exit "$status"
}
chmod +x "$cli_temp" "$runtime_temp"
"$cli_temp" version

mv -f "$runtime_temp" "$PORTER_HOME/runtimes/porter-runtime"
mv -f "$cli_temp" "$PORTER_HOME/porter"

"$PORTER_HOME/porter" mixin install exec --version "$PORTER_VERSION"

echo "Installation complete."
echo "Add porter to your path by adding the following line to your ~/.bash_profile or ~/.zprofile and open a new terminal:"
# Expand PATH when the user runs this command, not while installing.
# shellcheck disable=SC2016
printf 'export PATH="$PATH":%q\n' "$PORTER_HOME"
