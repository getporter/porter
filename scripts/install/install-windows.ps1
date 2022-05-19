# Installs the porter CLI for a single user.
param(
    [String]
    # The version of Porter to install, such as vX.Y.Z, latest or canary.
    $PORTER_PERMALINK='latest',
    [String]
    # DEPRECATED. Plugin and mixin versions are pinned to versions compatible with the v0.38 stable release
    $PKG_PERMALINK='latest',
    [String]
    # Location where Porter is installed (defaults to ~/.porter).
    $PORTER_HOME="$env:USERPROFILE\.porter",
    [String]
    # Base URL where Porter assets, such as binaries and atom feeds, are downloaded. This lets you setup an internal mirror.
    $PORTER_MIRROR="https://cdn.porter.sh")

echo "Installing porter@$PORTER_PERMALINK to $PORTER_HOME from $PORTER_MIRROR"

$env:PORTER_HOME=$PORTER_HOME
$env:PORTER_MIRROR=$PORTER_MIRROR
mkdir -f $PORTER_HOME/runtimes

(new-object System.Net.WebClient).DownloadFile("$PORTER_MIRROR/$PORTER_PERMALINK/porter-windows-amd64.exe", "$PORTER_HOME\porter.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_MIRROR/$PORTER_PERMALINK/porter-linux-amd64", "$PORTER_HOME\runtimes\porter-runtime")
echo "Installed $(& $PORTER_HOME\porter.exe version)"

& $PORTER_HOME/porter mixin install exec --version $PORTER_PERMALINK
& $PORTER_HOME/porter mixin install kubernetes --version v0.28.5
& $PORTER_HOME/porter mixin install helm --version v0.13.4
& $PORTER_HOME/porter mixin install arm --version v0.8.2
& $PORTER_HOME/porter mixin install terraform --version v0.9.1
& $PORTER_HOME/porter mixin install az --version v0.7.2
& $PORTER_HOME/porter mixin install aws --version v0.4.1
& $PORTER_HOME/porter mixin install gcloud --version v0.4.2

& $PORTER_HOME/porter plugin install azure --version v0.11.2

echo "Installation complete."
echo "Add porter to your path by adding the following line to your Microsoft.PowerShell_profile.ps1 and open a new terminal:"
echo '$env:PATH+=";$env:USERPROFILE\.porter"'
