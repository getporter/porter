param([String]$PORTER_PERMALINK='latest', [String]$PKG_PERMALINK='latest')

$PORTER_HOME="$env:USERPROFILE\.porter"
$PORTER_URL="https://cdn.porter.sh"

echo "Installing porter to $PORTER_HOME"

mkdir -f $PORTER_HOME

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_PERMALINK/porter-windows-amd64.exe", "$PORTER_HOME\porter.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_PERMALINK/porter-linux-amd64", "$PORTER_HOME\porter-runtime")
echo "Installed $(& $PORTER_HOME\porter.exe version)"

& $PORTER_HOME/porter mixin install exec --version $PKG_PERMALINK
& $PORTER_HOME/porter mixin install kubernetes --version $PKG_PERMALINK
& $PORTER_HOME/porter mixin install helm --version $PKG_PERMALINK
& $PORTER_HOME/porter mixin install arm --version $PKG_PERMALINK
& $PORTER_HOME/porter mixin install terraform --version $PKG_PERMALINK
& $PORTER_HOME/porter mixin install az --version $PKG_PERMALINK
& $PORTER_HOME/porter mixin install aws --version $PKG_PERMALINK
& $PORTER_HOME/porter mixin install gcloud --version $PKG_PERMALINK

& $PORTER_HOME/porter plugin install azure --version $PKG_PERMALINK

echo "Installation complete."
echo "Add porter to your path by adding the following line to your Microsoft.PowerShell_profile.ps1 and open a new terminal:"
echo '$env:PATH+=";$env:USERPROFILE\.porter"'
