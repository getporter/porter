resource "random_string" "password" {
  length = 16
  special = true
  override_special = "/@£$"
}

resource "random_string" "name" {
  length = 5
  special = false
}

resource "azurerm_mysql_server" "bundle" {
  name                = var.server_name
  location            = var.location
  resource_group_name = var.backend_storage_resource_group

  sku_name = "B_Gen5_2"

  storage_profile {
    storage_mb            = 5120
    backup_retention_days = 7
    geo_redundant_backup  = "Disabled"
  }

  administrator_login          = var.mysql_admin
  administrator_login_password = random_string.password.result
  version                      = "5.7"
  ssl_enforcement              = "Disabled"
}

resource "azurerm_mysql_database" "bundle" {
  name                = var.database_name
  resource_group_name = var.backend_storage_resource_group
  server_name         = azurerm_mysql_server.bundle.name
  charset             = "utf8"
  collation           = "utf8_unicode_ci"
}
