package plugins

import "get.porter.sh/porter/pkg/plugins"

// PluginInterface for the data storage. This first part of the
// three-part plugin key is only seen/used by the plugins when the host is
// communicating with the plugin and is not exposed to users.
const PluginInterface = "storage"

// StoragePlugin is the interface used to wrap a storage plugin.
// It is not meant to be used directly.
type StoragePlugin interface {
	plugins.Plugin
	StorageProtocol
}
