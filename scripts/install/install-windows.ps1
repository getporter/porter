$PORTER_HOME="$env:USERPROFILE\.porter"
$PORTER_URL="https://deislabs.blob.core.windows.net/porter"
$PORTER_VERSION="UNKNOWN"
echo "Installing porter to $PORTER_HOME"

mkdir -force -p $PORTER_HOME/templates
mkdir -force -p $PORTER_HOME/mixins/porter
mkdir -force -p $PORTER_HOME/mixins/exec
mkdir -force -p $PORTER_HOME/mixins/helm
mkdir -force -p $PORTER_HOME/mixins/azure

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_VERSION/porter-windows-amd64.exe", "$PORTER_HOME\porter.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_VERSION/porter-runtime-linux-amd64", "$PORTER_HOME\mixins\porter\porter-runtime")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_VERSION/templates/porter.yaml", "$PORTER_HOME\templates\porter.yaml")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/$PORTER_VERSION/templates/run", "$PORTER_HOME\templates\run")
echo "Installed $(iex "$PORTER_HOME\porter.exe version")"

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/mixins/exec/$PORTER_VERSION/exec-windows-amd64.exe", "$PORTER_HOME\mixins\exec\exec.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/mixins/exec/$PORTER_VERSION/exec-runtime-linux-amd64", "$PORTER_HOME\mixins\exec\exec-runtime")
echo "Installed $(iex "$PORTER_HOME\mixins\exec\exec.exe version")"

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/mixins/helm/$PORTER_VERSION/helm-windows-amd64.exe", "$PORTER_HOME\mixins\helm\helm.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/mixins/helm/$PORTER_VERSION/helm-runtime-linux-amd64", "$PORTER_HOME\mixins\helm\helm-runtime")
echo "Installed $(iex "$PORTER_HOME\mixins\helm\helm.exe version")"

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/mixins/azure/$PORTER_VERSION/azure-windows-amd64.exe", "$PORTER_HOME\mixins\azure\azure.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/mixins/azure/$PORTER_VERSION/azure-runtime-linux-amd64", "$PORTER_HOME\mixins\azure\azure-runtime")
echo "Installed azure mixin"

echo "Installation complete. Add $PORTER_HOME to your PATH."
