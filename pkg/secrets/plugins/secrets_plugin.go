package plugins

import (
	"errors"

	"get.porter.sh/porter/pkg/plugins"
)

// PluginInterface for the secrets. This first part of the
// three-part plugin key is only seen/used by the plugins when the host is
// communicating with the plugin and is not exposed to users.
const PluginInterface = "secrets"

// ErrNotImplemented is the error to be returned if a method is not implemented
// in a secret plugin
var ErrNotImplemented = errors.New("not implemented")

// SecretsPlugin is the interface used to wrap a secrets plugin.
// It is not meant to be used directly.
type SecretsPlugin interface {
	plugins.Plugin
	SecretsProtocol
}
