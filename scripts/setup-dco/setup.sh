#!/usr/bin/env bash
set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
HOOK="$DIR/../../.git/hooks/prepare-commit-msg"

function install-hook {
    echo "Installing DCO git commit hook"
    cp $DIR/prepare-commit-msg $HOOK
}

if [[ -f $HOOK ]]; then
    read -p "The DCO hook is already installed. Ovewrite? [yN]" yn
    case $yn in
        [Yy]*)
            install-hook
            ;;
        *)
            echo "The DCO hook was not installed"
            ;;
    esac
else
    install-hook
fi

