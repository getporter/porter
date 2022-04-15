package plugins

import "errors"

const (
	// PluginInterface for secrets. This first part of the
	// three-part plugin key is only seen/used by the plugins when the host is
	// communicating with the plugin and is not exposed to users.
	PluginInterface = "secrets"

	// PluginProtocolVersion is the currently supported plugin protocol version for secrets.
	PluginProtocolVersion = 1
)

var (
	// ErrNotImplemented is the error to be returned if a method is not implemented
	// in a secret plugin
	ErrNotImplemented = errors.New("not implemented")
)
