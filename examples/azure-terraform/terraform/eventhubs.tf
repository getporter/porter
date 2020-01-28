resource "azurerm_eventhub_namespace" "hubs" {
  name                = "porterform-eventhub-ns"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  sku                 = "Standard"
  capacity            = 1

  tags = {
    environment = "Production"
  }
}

resource "azurerm_eventhub" "hubs" {
  name                = "porterform-eventhub"
  namespace_name      = azurerm_eventhub_namespace.hubs.name
  resource_group_name = azurerm_resource_group.rg.name
  partition_count     = 2
  message_retention   = 1
}