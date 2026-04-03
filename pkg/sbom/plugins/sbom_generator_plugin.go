package plugins

import "errors"

const (
	// PluginInterface for sbom generation. This first part of the
	// three-part plugin key is only seen/used by the plugins when the host is
	// communicating with the plugin and is not exposed to users.
	PluginInterface = "sbomGenerator"

	// PluginProtocolVersion is the currently supported plugin protocol version for sbom generation.
	PluginProtocolVersion = 1
)

// ErrNotImplemented is the error to be returned if a method is not implemented
// in an sbom generator plugin
var ErrNotImplemented = errors.New("not implemented")
