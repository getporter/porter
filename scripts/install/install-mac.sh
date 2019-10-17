#!/usr/bin/env bash
set -euo pipefail

PORTER_HOME=~/.porter
PORTER_URL=https://deislabs.blob.core.windows.net/porter
PORTER_PERMALINK=${PORTER_PERMALINK:-latest}
MIXIN_PERMALINK=${MIXIN_PERMALINK:-latest}
echo "Installing porter to $PORTER_HOME"

mkdir -p $PORTER_HOME

curl -fsSLo $PORTER_HOME/porter $PORTER_URL/$PORTER_PERMALINK/porter-darwin-amd64
curl -fsSLo $PORTER_HOME/porter-runtime $PORTER_URL/$PORTER_PERMALINK/porter-linux-amd64
chmod +x $PORTER_HOME/porter
chmod +x $PORTER_HOME/porter-runtime
echo Installed `$PORTER_HOME/porter version`

$PORTER_HOME/porter mixin install exec --version $MIXIN_PERMALINK
$PORTER_HOME/porter mixin install kubernetes --version $MIXIN_PERMALINK
$PORTER_HOME/porter mixin install helm --version $MIXIN_PERMALINK
$PORTER_HOME/porter mixin install azure --version $MIXIN_PERMALINK
$PORTER_HOME/porter mixin install terraform --version $MIXIN_PERMALINK
$PORTER_HOME/porter mixin install az --version $MIXIN_PERMALINK
$PORTER_HOME/porter mixin install aws --version $MIXIN_PERMALINK
$PORTER_HOME/porter mixin install gcloud --version $MIXIN_PERMALINK

echo "Installation complete."
echo "Add porter to your path by running:"
echo "export PATH=\$PATH:~/.porter"
