package plugins

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

// Plugin is a general interface for interacting with Porter plugins.
type Plugin interface {
	// Connect establishes a connection to the plugin.
	// Safe to call multiple times, the existing connection is reused.
	Connect() error

	// Close the connection to the plugin.
	// Save to call multiple times.
	Close() error
}

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
		return PluginKey{}, errors.New("invalid plugin key '%s', allowed format is [INTERFACE].BINARY.IMPLEMENTATION")
	}

	if key.Binary == "porter" {
		key.IsInternal = true
	}

	return key, nil
}
