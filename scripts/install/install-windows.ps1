param([String]$PORTER_PERMALINK='latest', [String]$MIXIN_PERMALINK='latest')

$PORTER_HOME="$env:USERPROFILE\.porter"
$PORTER_URL="https://cdn.deislabs.io/porter"

echo "Installing porter to $PORTER_HOME"

mkdir -f $PORTER_HOME

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_PERMALINK/porter-windows-amd64.exe", "$PORTER_HOME\porter.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_PERMALINK/porter-linux-amd64", "$PORTER_HOME\porter-runtime")
echo "Installed $(& $PORTER_HOME\porter.exe version)"

& $PORTER_HOME/porter mixin install exec --version $MIXIN_PERMALINK
& $PORTER_HOME/porter mixin install kubernetes --version $MIXIN_PERMALINK
& $PORTER_HOME/porter mixin install helm --version $MIXIN_PERMALINK
& $PORTER_HOME/porter mixin install azure --version $MIXIN_PERMALINK
& $PORTER_HOME/porter mixin install terraform --version $MIXIN_PERMALINK
& $PORTER_HOME/porter mixin install az --version $MIXIN_PERMALINK
& $PORTER_HOME/porter mixin install aws --version $MIXIN_PERMALINK
& $PORTER_HOME/porter mixin install gcloud --version $MIXIN_PERMALINK

echo "Installation complete."
echo "Add porter to your path by running:"
echo '$env:PATH+=";$env:USERPROFILE\.porter"'
