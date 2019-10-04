package plugins

import (
	"github.com/hashicorp/go-plugin"
)

// Common handshake config between Porter and its plugins.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PORTER",
	MagicCookieValue: "bbc2dd71-def4-4311-906e-e98dc27208ce",
}
