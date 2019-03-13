#!/usr/bin/env bash
set -euo pipefail

PORTER_HOME=~/.porter
PORTER_URL=https://deislabs.blob.core.windows.net/porter
PORTER_VERSION=${PORTER_VERSION:-UNKNOWN}
echo "Installing porter to $PORTER_HOME"

mkdir -p $PORTER_HOME/templates
mkdir -p $PORTER_HOME/mixins/porter
mkdir -p $PORTER_HOME/mixins/exec
mkdir -p $PORTER_HOME/mixins/helm
mkdir -p $PORTER_HOME/mixins/azure

curl -fsSLo $PORTER_HOME/porter $PORTER_URL/$ORTER_VERSION/porter-linux-amd64
curl -fsSLo $PORTER_HOME/mixins/porter/porter-runtime $PORTER_URL/$PORTER_VERSION/porter-runtime-linux-amd64
curl -fsSLo $PORTER_HOME/templates/porter.yaml $PORTER_URL/$PORTER_VERSION/templates/porter.yaml
curl -fsSLo $PORTER_HOME/templates/run $PORTER_URL/$PORTER_VERSION/templates/run
chmod +x $PORTER_HOME/porter
chmod +x $PORTER_HOME/mixins/porter/porter-runtime
echo Installed `$PORTER_HOME/porter version`

curl -fsSLo $PORTER_HOME/mixins/exec/exec $PORTER_URL/mixins/exec/$PORTER_VERSION/exec-linux-amd64
curl -fsSLo $PORTER_HOME/mixins/exec/exec-runtime $PORTER_URL/mixins/exec/$PORTER_VERSION/exec-runtime-linux-amd64
chmod +x $PORTER_HOME/mixins/exec/exec
chmod +x $PORTER_HOME/mixins/exec/exec-runtime
echo Installed `$PORTER_HOME/mixins/exec/exec version`

curl -fsSLo $PORTER_HOME/mixins/helm/helm $PORTER_URL/mixins/helm/$PORTER_VERSION/helm-linux-amd64
curl -fsSLo $PORTER_HOME/mixins/helm/helm-runtime $PORTER_URL/mixins/helm/$PORTER_VERSION/helm-runtime-linux-amd64
chmod +x $PORTER_HOME/mixins/helm/helm
chmod +x $PORTER_HOME/mixins/helm/helm-runtime
echo Installed `$PORTER_HOME/mixins/helm/helm version`

curl -fsSLo $PORTER_HOME/mixins/azure/azure $PORTER_URL/mixins/azure/$PORTER_VERSION/azure-linux-amd64
curl -fsSLo $PORTER_HOME/mixins/azure/azure-runtime $PORTER_URL/mixins/azure/$PORTER_VERSION/azure-runtime-linux-amd64
chmod +x $PORTER_HOME/mixins/azure/azure
chmod +x $PORTER_HOME/mixins/azure/azure-runtime
echo Installed azure mixin

echo "Installation complete. Add $PORTER_HOME to your PATH."
