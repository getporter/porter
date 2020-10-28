variable "client_id" {}
variable "client_secret" {}
variable "tenant_id" {}
variable "subscription_id" {}

variable "location" {
    default = "EastUS"
}

variable "backend_storage_account" {}

variable "backend_storage_resource_group" {
    default = "devops-days-msp"

}
variable "backend_storage_container" { 
    default = "tf-storage"
}

variable "server-name" {
    default = "mysql-bundle"
}

variable "mysql-admin" {
   default = "myadmin"
}

variable "database-name" {
   default = "workshop"
}