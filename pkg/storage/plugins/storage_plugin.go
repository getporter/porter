package plugins

const (
	// PluginInterface for storage. This first part of the
	// three-part plugin key is only seen/used by the plugins when the host is
	// communicating with the plugin and is not exposed to users.
	PluginInterface = "storage"

	// PluginProtocolVersion is the currently supported plugin protocol version for storage.
	PluginProtocolVersion = 3
)
