$PORTER_HOME="$env:USERPROFILE\.porter"
$PORTER_URL="https://cdn.deislabs.io/porter"
$PORTER_VERSION="latest"
echo "Installing porter to $PORTER_HOME"

mkdir -f $PORTER_HOME

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_VERSION/porter-windows-amd64.exe", "$PORTER_HOME\porter.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_VERSION/porter-linux-amd64", "$PORTER_HOME\porter-runtime")
echo "Installed $(iex "$PORTER_HOME\porter.exe version")"

$FEED_URL="$PORTER_URL/atom.xml"
iex "$PORTER_HOME/porter mixin install exec --version $PORTER_VERSION --feed-url $FEED_URL"
iex "$PORTER_HOME/porter mixin install kubernetes --version $PORTER_VERSION --feed-url $FEED_URL"
iex "$PORTER_HOME/porter mixin install helm --version $PORTER_VERSION --feed-url $FEED_URL"
iex "$PORTER_HOME/porter mixin install azure --version $PORTER_VERSION --feed-url $FEED_URL"

echo "Installation complete."
echo "Add porter to your path by running:"
echo '$env:PATH+=";$env:USERPROFILE\.porter"'
