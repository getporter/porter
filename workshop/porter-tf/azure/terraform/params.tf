variable "client_id" {}
variable "client_secret" {}
variable "tenant_id" {}
variable "subscription_id" {}

variable "location" {
    default = "EastUS"
}

variable "backend_storage_resource_group" {
    default = "devops-days-msp"
}

variable "server_name" {
    default = "mysql-bundle"
}

variable "mysql_admin" {
   default = "myadmin"
}

variable "database_name" {
   default = "workshop"
}