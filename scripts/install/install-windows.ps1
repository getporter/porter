$PORTER_HOME="$env:USERPROFILE\.porter"
$PORTER_URL="https://deislabs.blob.core.windows.net/porter"
$PORTER_VERSION="UNKNOWN"
echo "Installing porter to $PORTER_HOME"

mkdir -f $PORTER_HOME

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_VERSION/porter-windows-amd64.exe", "$PORTER_HOME\porter.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_VERSION/porter-linux-amd64", "$PORTER_HOME\porter-runtime")
echo "Installed $(iex "$PORTER_HOME\porter.exe version")"

$MIXINS_URL="$PORTER_URL/mixins"
iex "$PORTER_HOME/porter mixin install exec --version $PORTER_VERSION --url $MIXINS_URL/exec"
iex "$PORTER_HOME/porter mixin install helm --version $PORTER_VERSION --url $MIXINS_URL/helm"
iex "$PORTER_HOME/porter mixin install azure --version $PORTER_VERSION --url $MIXINS_URL/azure"

echo "Installation complete."
echo "Add porter to your path by running:"
echo '$env:PATH+=";$env:USERPROFILE\.porter"'
