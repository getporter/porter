#!/usr/bin/env bash
set -euo pipefail

PORTER_HOME=~/.porter
PORTER_URL=https://cdn.porter.sh
PORTER_PERMALINK=${PORTER_PERMALINK:-latest}
PKG_PERMALINK=${PKG_PERMALINK:-latest}
echo "Installing porter to $PORTER_HOME"

mkdir -p $PORTER_HOME/runtimes

curl --http1.1 -fsSLo $PORTER_HOME/porter $PORTER_URL/$PORTER_PERMALINK/porter-linux-amd64
chmod +x $PORTER_HOME/porter
cp $PORTER_HOME/porter $PORTER_HOME/runtimes/porter-runtime
echo Installed `$PORTER_HOME/porter version`

$PORTER_HOME/porter mixin install exec --version $PKG_PERMALINK
$PORTER_HOME/porter mixin install kubernetes --version $PKG_PERMALINK
$PORTER_HOME/porter mixin install helm --version $PKG_PERMALINK
$PORTER_HOME/porter mixin install arm --version $PKG_PERMALINK
$PORTER_HOME/porter mixin install terraform --version $PKG_PERMALINK
$PORTER_HOME/porter mixin install az --version $PKG_PERMALINK
$PORTER_HOME/porter mixin install aws --version $PKG_PERMALINK
$PORTER_HOME/porter mixin install gcloud --version $PKG_PERMALINK

$PORTER_HOME/porter plugin install azure --version $PKG_PERMALINK

echo "Installation complete."
echo "Add porter to your path by adding the following line to your ~/.profile and open a new terminal:"
echo "export PATH=\$PATH:~/.porter"
