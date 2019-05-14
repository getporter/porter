# CNAB Bundle with Porter - Spring Music Demo App

This project is a sample CNAB bundle that is created and managed by Porter. https://porter.sh 

This bundle uses 3 mixins to access your Azure subscription and deploy the app. These values need to be updated in the porter.yaml.

* The `exec` mixin uses an Azure Service Principal to access via the CLI.
* The `azure` mixin uses ARM and requires the subscription and tenant info. 
* The `helm` mixin deploys the chart into your AKS cluster.

### Prerequisites

- Porter on local machine. http://porter.sh
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
    * change the `invocationImage` Docker repo to match your Docker Hub account (line 4)
    * Cosmos and AKS names must be unique. You can either edit the `porter.yaml` file default values (starting on line 90) or you can supply the with the porter CLI as shown below.

* Build the innvocation image

    ```bash
    porter build
    ```

* Install the bundle

    ```bash
    export INSTALL_ID=314
    porter install -c azure  --param app-resource-group=spring-music-demo-$INSTALL_ID --param aks-resource-group=spring-music-demo-$INSTALL_ID --param aks-cluster-name=briar-aks-spring-$INSTALL_ID --param cosmosdb-service-name=briarspringmusic$INSTALL_ID --param azure-location=eastus
    ```
