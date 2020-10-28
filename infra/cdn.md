# Porter's CDN Setup

Porter uses a variety of services to distribute its CLI, mixin and plugin binaries.

## Azure Storage

Binaries are uploaded to [Azure Blob Storage](https://azure.microsoft.com/en-us/services/storage/blobs/) upon relevant events in our GitHub repositories, such as merges to the "main" branch (producing `canary` artifacts) and official releases (producing `latest` and semver-tagged artifacts).

Although it is possible to provide URLs directly to the stored resources, we'd be tightly coupled to a particular storage account and layout, not to mention ungainly URLs.  Therefore, we utilize the services below to achieve flexibility and control over asset links.

## DNS for porter.sh

DNS entries for the `porter.sh` domain are managed via a [Netlify](https://www.netlify.com/) account.  A `cdn` CNAME record exists in this configuration such that `cdn.porter.sh` can represent the hostname for all artifact URLs.

## Azure Front Door

We use [Azure Front Door](https://azure.microsoft.com/en-us/services/frontdoor/) to route incoming requests involving the `https://cdn.porter.sh` hostname to their corresponding assets in storage.  This is where TLS certificate material for our custom domain is managed, along with routing rules corresponding to various classes of assets (mixins, plugins, etc.)

## Azure Web Application Firewall

As a first stop for all incoming traffic, we use [Azure Web Application Firewall](https://docs.microsoft.com/en-us/azure/web-application-firewall/) to filter out requests involving invalid URIs.  This ensures that most of the traffic actually reaching Azure Front Door corresponds to actual assets.