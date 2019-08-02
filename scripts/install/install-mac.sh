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

FEED_URL=$PORTER_URL/atom.xml
$PORTER_HOME/porter mixin install exec --version $PORTER_VERSION --feed-url $FEED_URL
$PORTER_HOME/porter mixin install kubernetes --version $PORTER_VERSION --feed-url $FEED_URL
$PORTER_HOME/porter mixin install helm --version $PORTER_VERSION --feed-url $FEED_URL
$PORTER_HOME/porter mixin install azure --version $PORTER_VERSION --feed-url $FEED_URL
$PORTER_HOME/porter mixin install terraform --version $PORTER_VERSION --feed-url $FEED_URL

echo "Installation complete."
echo "Add porter to your path by running:"
echo "export PATH=\$PATH:~/.porter"
