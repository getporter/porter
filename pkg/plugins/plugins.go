package plugins

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-plugin"
)

// Common handshake config between Porter and its plugins.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
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
		key.IsInternal = true
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
		return PluginKey{}, errors.New("invalid plugin key %q, allowed format is [INTERFACE].BINARY.IMPLEMENTATION")
	}

	return key, nil
}
