# CNAB Bundle with Porter - Spring Music Demo App

This bundle demonstrates advanced use cases for Porter.

The bundle leverages a base Dockerfile (Dockerfile.tmpl) to customize the resulting invocation image for the bundle by first installing the `azure cli` so that it can be used by the `exec` mixin. It then uses 4 mixins to access your Azure subscription and deploy the app. These values need to be updated in the porter.yaml.

* The `arm` mixin is used to create an AKS cluster using ARM. This requires subscription and tenant info.
* The `exec` mixin uses an Azure Service Principal to access via the CLI and install Helm's Tiller into an AKS cluster.
* The `kubernetes` mixin applys RBAC policies for Helm
* The `helm` mixin deploys the chart into the AKS cluster.


### Prerequisites

- Porter on local machine. See these helpful [installation instructions](https://porter.sh/install) 
- Docker on local machine (eg - Docker for Mac)
- Bash
- Azure service principal with rights to create a RG, AKS, Cosmos, etc. 

    ```bash
    az ad sp create-for-rbac --name ServicePrincipalName
    ```

    More details here: https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest 

- Follow the process in the porter docs to add these credentials in your config.


### Build / Install this bundle

* Setup credentials with Porter

The bundle will use the service principal created above to interact with Azure. Generate a credential using the `porter credentials generate` command:

    ```bash
    porter credentials generate azure 
    ```

* Update params for your deployment
    * change the `tag` Docker repo to match your Docker Hub account
    * Cosmos and AKS names must be unique. You can either edit the `porter.yaml` file default values (starting on line 90) or you can supply the with the porter CLI as shown below.

* Build the innvocation image

    ```bash
    porter build
    ```

* Install the bundle

    ```bash
    export INSTALL_ID=314
    porter install -c azure  \
      --param app-resource-group=spring-music-demo-$INSTALL_ID \
      --param aks-resource-group=spring-music-demo-$INSTALL_ID \
      --param aks-cluster-name=briar-aks-spring-$INSTALL_ID \
      --param cosmosdb-service-name=briarspringmusic$INSTALL_ID \
      --param azure-location=eastus
    ```
