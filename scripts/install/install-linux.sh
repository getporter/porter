#!/usr/bin/env bash
set -euo pipefail

# Installs the porter CLI for a single user.
# PORTER_HOME:      Location where Porter is installed (defaults to ~/.porter).
# PORTER_MIRROR:       Base URL where Porter assets, such as binaries and atom feeds, are downloaded. This lets you
#                   setup an internal mirror.
# PORTER_PERMALINK: The version of Porter to install, such as vX.Y.Z, latest or canary.
# PKG_PERMALINK:    DEPRECATED. Plugin and mixin versions are pinned to versions compatible with the v0.38 stable release

export PORTER_HOME=${PORTER_HOME:-~/.porter}
export PORTER_MIRROR=${PORTER_MIRROR:-https://cdn.porter.sh}
PORTER_PERMALINK=${PORTER_PERMALINK:-latest}

echo "Installing porter@$PORTER_PERMALINK to $PORTER_HOME from $PORTER_MIRROR"

mkdir -p $PORTER_HOME/runtimes

curl -fsSLo $PORTER_HOME/porter $PORTER_MIRROR/$PORTER_PERMALINK/porter-linux-amd64
chmod +x $PORTER_HOME/porter
cp $PORTER_HOME/porter $PORTER_HOME/runtimes/porter-runtime
echo Installed `$PORTER_HOME/porter version`

$PORTER_HOME/porter mixin install exec --version $PORTER_PERMALINK
$PORTER_HOME/porter mixin install kubernetes --version v0.28.5
$PORTER_HOME/porter mixin install helm --version v0.13.4
$PORTER_HOME/porter mixin install arm --version v0.8.2
$PORTER_HOME/porter mixin install terraform --version v0.9.1
$PORTER_HOME/porter mixin install az --version v0.7.2
$PORTER_HOME/porter mixin install aws --version v0.4.1
$PORTER_HOME/porter mixin install gcloud --version v0.4.2

$PORTER_HOME/porter plugin install azure --version v0.11.2

echo "Installation complete."
echo "Add porter to your path by adding the following line to your ~/.profile and open a new terminal:"
echo "export PATH=\$PATH:~/.porter"
