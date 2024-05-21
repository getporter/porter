package plugins

import "context"

// SigningProtocol is the interface that signing plugins must implement.
// This defines the protocol used to communicate with signing plugins.
type SigningProtocol interface {
	Connect(ctx context.Context) error

	// Resolve a secret value from a secret store
	// - ref is OCI reference to verify
	Sign(ctx context.Context, ref string) error

	// Verify attempts to verify the signature of a given reference
	// - ref is OCI reference to verify
	Verify(ctx context.Context, ref string) error
}
