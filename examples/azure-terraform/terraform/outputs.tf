output "cosmos-db-uri" {
  value = azurerm_cosmosdb_account.db.connection_strings[0]
  sensitive = true
}

output "eventhubs_connection_string" {
  value = azurerm_eventhub_namespace.hubs.default_primary_connection_string
  sensitive = true
}

output "eventhubs_topic" {
  value = azurerm_eventhub.hubs.name
}