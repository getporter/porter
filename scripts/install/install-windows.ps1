$PORTER_HOME="$env:USERPROFILE\.porter"
$PORTER_URL="https://cdn.deislabs.io/porter"
$PORTER_VERSION="latest"
echo "Installing porter to $PORTER_HOME"

mkdir -f $PORTER_HOME

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_VERSION/porter-windows-amd64.exe", "$PORTER_HOME\porter.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_VERSION/porter-linux-amd64", "$PORTER_HOME\porter-runtime")
echo "Installed $(& $PORTER_HOME\porter.exe version)"

& $PORTER_HOME/porter mixin install exec --version $PORTER_VERSION
& $PORTER_HOME/porter mixin install kubernetes --version $PORTER_VERSION
& $PORTER_HOME/porter mixin install helm --version $PORTER_VERSION
& $PORTER_HOME/porter mixin install azure --version $PORTER_VERSION
& $PORTER_HOME/porter mixin install terraform --version $PORTER_VERSION

echo "Installation complete."
echo "Add porter to your path by running:"
echo '$env:PATH+=";$env:USERPROFILE\.porter"'
