variable "client_id" {}
variable "client_secret" {}
variable "tenant_id" {}
variable "subscription_id" {}


variable "database_name" {}

variable "resource_group_name" {
    default = "azure-porter-tf"
}

variable "resource_group_location" {
    default = "East US"
}

variable "failover_location" {
    default = "West US"
}