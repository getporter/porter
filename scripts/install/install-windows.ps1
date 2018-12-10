$PORTER_HOME="$env:USERPROFILE\.porter"
$PORTER_URL="https://deislabs.blob.core.windows.net/porter"
echo "Installing porter to $PORTER_HOME"

mkdir -p $PORTER_HOME/templates
mkdir -p $PORTER_HOME/mixins/porter
mkdir -p $PORTER_HOME/mixins/exec
mkdir -p $PORTER_HOME/mixins/helm
mkdir -p $PORTER_HOME/mixins/azure

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/latest/porter-windows-amd64.exe", "$PORTER_HOME\porter.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/latest/porter-runtime-linux-amd64", "$PORTER_HOME\mixins\porter\porter-runtime")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/latest/templates/porter.yaml", "$PORTER_HOME\templates\porter.yaml")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/latest/templates/run", "$PORTER_HOME\templates\run")
echo "Installed $(iex "$PORTER_HOME\porter.exe version")"

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/mixins/exec/latest/exec-windows-amd64.exe", "$PORTER_HOME\mixins\exec\exec.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/mixins/exec/latest/exec-runtime-linux-amd64", "$PORTER_HOME\mixins\exec\exec-runtime.exe")
echo "Installed $(iex "$PORTER_HOME\mixins\exec\exec.exe version")"

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/mixins/helm/latest/helm-windows-amd64.exe", "$PORTER_HOME\mixins\helm\helm.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/mixins/helm/latest/helm-runtime-linux-amd64", "$PORTER_HOME\mixins\helm\helm-runtime.exe")
echo "Installed $(iex "$PORTER_HOME\mixins\helm\helm.exe version")"

(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/mixins/azure/latest/azure-windows-amd64.exe", "$PORTER_HOME\mixins\azure\azure.exe")
(new-object System.Net.WebClient).DownloadFile("$PORTER_URL/mixins/azure/latest/azure-runtime-linux-amd64", "$PORTER_HOME\mixins\azure\azure-runtime.exe")
echo "Installed azure mixin"

echo "Installation complete. Add $PORTER_HOME to your PATH."
