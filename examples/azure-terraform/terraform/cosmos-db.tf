resource "azurerm_resource_group" "rg" {
  name     = var.resource_group_name
  location = var.resource_group_location
}

resource "azurerm_cosmosdb_account" "db" {
  name                = "porterform-cosmos-db"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  offer_type          = "Standard"
  kind                = "MongoDB"

  enable_automatic_failover = false

  consistency_policy {
    consistency_level       = "BoundedStaleness"
    max_interval_in_seconds = 301
    max_staleness_prefix    = 100001
  }

  geo_location {
    location          = var.failover_location
    failover_priority = 1
  }

  geo_location {
    prefix            = "porterform-${azurerm_resource_group.rg.location}"
    location          = azurerm_resource_group.rg.location
    failover_priority = 0
  }
}

resource "azurerm_cosmosdb_mongo_database" "db" {
  name                = var.database_name
  resource_group_name = azurerm_cosmosdb_account.db.resource_group_name
  account_name        = azurerm_cosmosdb_account.db.name
}