---
title: Connect to AKS
description: How to connect to an AKS cluster inside a Porter bundle.
weight: 1
---

* AKS cluster authentication requires more than just a kubeconfig.

### Prerequisites

* To access resources that are secured by a Microsoft Entra tenant, the entity that requires access must be represented by a security principal. This requirement is true for both users (user principal) and applications (service principal). 

* Furthermore, the following will be used:
  * A **service principal** which is the local representation of an application object in a single Microsoft Entra tenant. 
  * A **client secret** used for authenticating with the service principal.

* You can either create a new service principal (az cli guide [here](https://learn.microsoft.com/en-us/cli/azure/ad/sp?view=azure-cli-latest#az-ad-sp-create)) or use an existing one.

* From Azure portal: Microsoft Entra admin center (under App registrations) select your application and get: `Application (client) ID`, `Object ID`, `Directory (tenant) ID` and set them as env variables.

```bash
export AZURE_CLIENT_ID="insert_ApplicationID"
export AZURE_TENANT_ID="insert_TenantID"
export AZURE_CLIENT_SECRET="insert_SecretValue"
export AZURE_SUBSCRIPTION_ID="insert_SubscriptionID"
```

### Bundle setup

* After the variables have been set, create a credentials set, by saving the following json to a file e.g. `aks_creds.json` and apply the file: `porter credentials apply aks_creds.json`

```json
{
    "schemaType": "CredentialSet",
    "schemaVersion": "1.0.1",
    "name": "akscreds",
    "namespace": "",
    "credentials": [

        {
            "name": "azure_client_id",
            "source": {
                "env": "AZURE_CLIENT_ID"
            }
        },
        
        {
            "name": "azure_client_secret",
            "source": {
                "env": "AZURE_CLIENT_SECRET"
            }
        },

        {
            "name": "azure_subscription_id",
            "source": {
                "env": "AZURE_SUBSCRIPTION_ID"
            }
        },

        {
            "name": "azure_tenant_id",
            "source": {
                "env": "AZURE_TENANT_ID"
            }
        }
    ]
}
```

* Verify that `akscreds` credentials set exists: `porter credentials list`

* Install [az mixin](https://porter.sh/mixins/az/) and declare the credentials in the bundle manifest:

```yaml
credentials:
- name: azure_client_id
  env: AZURE_CLIENT_ID
  description: AAD Client ID for Azure account authentication - used for AKS Cluster SPN details and for authentication to azure to get KubeConfig
- name: azure_tenant_id
  env: AZURE_TENANT_ID
  description: Azure AAD Tenant Id for Azure account authentication - used to authenticate to Azure to get KubeConfig 
- name: azure_client_secret
  env: AZURE_CLIENT_SECRET
  description: AAD Client Secret for Azure account authentication - used for AKS Cluster SPN details and for authentication to azure to get KubeConfig
- name: azure_subscription_id
  env: AZURE_SUBSCRIPTION_ID
  description: Azure Subscription Id used to set the subscription where the account has access to multiple subscriptions
```

* Declare Azure resource group and AKS name as parameters:
```yaml
parameters:
  - name: rg_name
    type: string
    description: "Resource group in which AKS resides"
    default: "demorg"
    applyTo:
      - merge
      
  - name: aks_name
    type: string
    description: "AKS name"
    default: "demoaks"
    applyTo:
      - merge
```

* Bellow you can find a custom action called `merge` that it's used to autheticate in Azure and sets the desired subscription (that contains the resource group and AKS cluster) and uses [outputs](https://porter.sh/wiring/#outputs) to save the *kubeconfig* file:

```bash
merge:
  - az:
      description: "Azure login"
      arguments:
        - login
      flags:
        service-principal:
        username: ${ bundle.credentials.azure_client_id }
        password: ${ bundle.credentials.azure_client_secret }
        tenant: ${ bundle.credentials.azure_tenant_id }
  - az:
      description: "Azure set subscription Id"
      arguments:
        - "account"
        - "set"
      flags:
        subscription: ${ bundle.credentials.azure_subscription_id }
  - az:
      description: "Get access creds for AKS"
      arguments:
        - "aks"
        - "get-credentials"
      flags:
        resource-group: ${ bundle.parameters.rg_name }
        name: ${ bundle.parameters.aks_name }
        context: akscluster
        file: "newconf"
      suppress-output: false
      outputs:
        - name: kubeconfig
          path: newconf

outputs:
  - name: newconf
    type: file
    path: newconf
```

* Write the *kubeconfig* file to disk: `porter installations output show newconf > aks_config`, afterwards you can use it as it is: `kubectl --kubeconfig=aks_config config get-contexts` or merge it in the existing `~/.kube/config`.