package signing

import "context"

// Signer is an interface for signing and verifying Porter
// bundles and bundle images.
type Signer interface {
	Close() error

	// Sign generates a signature for the specified reference, which
	// can be a Porter bundle or an bundle image.
	Sign(ctx context.Context, ref string) error
	// Verify attempts to verify a signature for the specified
	// reference, which can be a Porter bundle or an bundle image.
	Verify(ctx context.Context, ref string) error
	// TODO
	Connect(ctx context.Context) error
}
