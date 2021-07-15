package pluggable

import (
	"get.porter.sh/porter/pkg/config"
	"github.com/hashicorp/go-plugin"
)

// Entry defines a configuration entry for an item that can be managed by a plugin.
type Entry interface {
	GetName() string
	GetPluginSubKey() string
	GetConfig() interface{}
}

// PluginTypeConfig defines a set of functions to access a type of plugin's data in
// the porter config file.
type PluginTypeConfig struct {
	// Name of the plugin type interface.
	Interface string

	// Plugin to communicate with the plugin
	Plugin plugin.Plugin

	// GetDefaultPluggable is the function on porter's configuration
	// to retrieve a pluggable configuration value's named default instance to use, e.g. "default-storage"
	GetDefaultPluggable func(c *config.Config) string

	// GetPluggable is the function on porter's configuration
	// to retrieve a named pluggable instance, e.g. a storage named "azure"
	GetPluggable func(c *config.Config, name string) (Entry, error)

	// GetDefaultPlugin is the function on porter's configuration
	// to retrieve the default plugin to use for a type of plugin, e.g. "storage-plugin"
	GetDefaultPlugin func(c *config.Config) string

	// ProtocolVersion is the version of the protocol used by this plugin.
	ProtocolVersion uint
}
