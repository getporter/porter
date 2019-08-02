resource "random_string" "password" {
  length = 16
  special = true
  override_special = "/@\" "
}

resource "random_string" "name" {
  length = 5
   special = false
}

resource "azurerm_mysql_server" "bundle" {
  name                = "${var.server-name}"
  location            = "EastUS"
  resource_group_name = "devops-days-msp"

  sku {
    name     = "B_Gen5_2"
    capacity = 2
    tier     = "Basic"
    family   = "Gen5"
  }

  storage_profile {
    storage_mb            = 5120
    backup_retention_days = 7
    geo_redundant_backup  = "Disabled"
  }

  administrator_login          = "${var.mysql-admin}"
  administrator_login_password = "${random_string.password.result}"
  version                      = "5.7"
  ssl_enforcement              = "Disabled"
}

resource "azurerm_mysql_database" "bundle" {
  name                = "${var.database_name}"
  resource_group_name = "devops-days-msp"
  server_name         = "${azurerm_mysql_server.bundle.name}"
  charset             = "utf8"
  collation           = "utf8_unicode_ci"
}
