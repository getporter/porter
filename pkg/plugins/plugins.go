package plugins

import (
	"context"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"github.com/hashicorp/go-plugin"
)

// HandshakeConfig is common handshake config between Porter and its plugins.
var HandshakeConfig = plugin.HandshakeConfig{
	MagicCookieKey:   "PORTER",
	MagicCookieValue: "bbc2dd71-def4-4311-906e-e98dc27208ce",
}

type PluginKey struct {
	Binary         string
	Interface      string
	Implementation string
	IsInternal     bool
}

func (k PluginKey) String() string {
	return fmt.Sprintf("%s.%s.%s", k.Interface, k.Binary, k.Implementation)
}

func ParsePluginKey(value string) (PluginKey, error) {
	var key PluginKey

	parts := strings.Split(value, ".")

	switch len(parts) {
	case 1:
		key.Binary = "porter"
		key.Implementation = parts[0]
	case 2:
		key.Binary = parts[0]
		key.Implementation = parts[1]
	case 3:
		key.Interface = parts[0]
		key.Binary = parts[1]
		key.Implementation = parts[2]
	default:
		return PluginKey{}, fmt.Errorf("invalid plugin key '%s', allowed format is [INTERFACE].BINARY.IMPLEMENTATION", value)
	}

	if key.Binary == "porter" {
		key.IsInternal = true
	}

	return key, nil
}

// PluginRegistration is the info needed to automatically handle
// running a plugin when requested.
type PluginRegistration struct {
	// Interface that the plugin implements, such as storage or secrets.
	Interface string

	// ProtocolVersion is the version of the plugin protocol that the plugin supports.
	ProtocolVersion int

	// Create is the handler called to make an instance of the plugin.
	Create func(c *config.Config, pluginCfg interface{}) (plugin.Plugin, error)
}

// PluginCloser is the interface that plugins should implement when they need to
// clean up resources when Porter is done with the plugin.
type PluginCloser interface {
	// Close requests that the plugin clean up long-held resources.
	// A context is passed so that the plugin can still output log/trace data.
	Close(ctx context.Context) error
}
