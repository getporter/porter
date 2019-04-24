#!/usr/bin/env bash
set -euo pipefail

PORTER_HOME=~/.porter
PORTER_URL=https://deislabs.blob.core.windows.net/porter
PORTER_VERSION=${PORTER_VERSION:-UNKNOWN}
echo "Installing porter to $PORTER_HOME"

mkdir -p $PORTER_HOME

curl -fsSLo $PORTER_HOME/porter $PORTER_URL/$PORTER_VERSION/porter-darwin-amd64
curl -fsSLo $PORTER_HOME/porter-runtime $PORTER_URL/$PORTER_VERSION/porter-linux-amd64
chmod +x $PORTER_HOME/porter
chmod +x $PORTER_HOME/porter-runtime
echo Installed `$PORTER_HOME/porter version`

MIXINS_URL=$PORTER_URL/mixins
$PORTER_HOME/porter mixin install exec --version $PORTER_VERSION --url $MIXINS_URL/exec
$PORTER_HOME/porter mixin install helm --version $PORTER_VERSION --url $MIXINS_URL/helm
$PORTER_HOME/porter mixin install azure --version $PORTER_VERSION --url $MIXINS_URL/azure

echo "Installation complete."
echo "Add porter to your path by running:"
echo "export PATH=\$PATH:~/.porter"
