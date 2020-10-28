---
title: Personas
descriptions: Know your users!
---

## Helpful Application Bundle Author - Liz
<img src="/images/personas/nerdy-lady.png" width="250" />

This is a helm chart author who also wants to make their chart available as a
bundle for people to try out. For example, the author of the MySQL or MongoDB charts.

They only run their bundle locally, or on a single cloud environment, just to
make sure that it works.

## Enterprising Vendor Bundle Author - Chen
<img src="/images/personas/neckbeard-gopher.png" width="250" />
This is a vendor, such as Azure, who also wants to make their cloud service available
as a bundle for people to try out. For example, the CosmosDB team wants customers
to use Cosmos as a bundle.

They only run their bundle on their own cloud, it's not super configurable or portable.

## Plucky In-House Developer - Peter
<img src="/images/personas/business-gopher.png" width="250" />
This is a developer of a multi-component suite of software, who uses application
bundles to satisfy dependencies. For example, the author of a boring line of business
application.

They need to be able to work with the bundle in their local development environment,
and then hand it off to another person to deploy it in the test and production environments.

## Beleaguered Cluster Operator - Tamara
<img src="/images/personas/quirky-gopher.png" width="250" />
This is the person in operations who is handed a bundle to deploy in their CI pipeline.
They are responsible for deploying in multiple environments, such as test and production.

They need to configure and deploy the same bundle into different environments which
run on different platforms and use different underlying service providers. For example,
their test environment uses an on-premise Kubernetes cluster and a mega shared
MS SQL server, while their production environment uses AKS and Azure SQL Server.
