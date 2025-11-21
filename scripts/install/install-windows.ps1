# Installs the porter CLI for a single user.
param(
    [String]
    # The version of Porter to install, such as vX.Y.Z, latest or canary.
    $PORTER_VERSION='latest',
    [String]
    # Location where Porter is installed (defaults to ~/.porter).
    $PORTER_HOME="$env:USERPROFILE\.porter",
    [String]
    # Base URL where Porter assets, such as binaries and atom feeds, are downloaded. This lets you setup an internal mirror.
    $PORTER_MIRROR="https://cdn.porter.sh")

echo "Installing porter@$PORTER_VERSION to $PORTER_HOME from $PORTER_MIRROR"

$env:PORTER_HOME=$PORTER_HOME
$env:PORTER_MIRROR=$PORTER_MIRROR
mkdir -f $PORTER_HOME/runtimes

(new-object System.Net.WebClient).DownloadFile("$PORTER_MIRROR/$PORTER_VERSION/porter-windows-amd64.exe", "$PORTER_HOME\porter.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_MIRROR/$PORTER_VERSION/porter-linux-amd64", "$PORTER_HOME\runtimes\porter-runtime")
echo "Installed $(& $PORTER_HOME\porter.exe version)"

& $PORTER_HOME/porter mixin install exec --version $PORTER_VERSION

echo "Installation complete."
echo "Add porter to your path by adding the following line to your Microsoft.PowerShell_profile.ps1 and open a new terminal:"
echo '$env:PATH+=";$env:USERPROFILE\.porter"'
