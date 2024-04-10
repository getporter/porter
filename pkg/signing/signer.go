package signing

import "context"

// Signer is an interface for signing and verifying Porter
// bundles and invocation images.
type Signer interface {
	// Sign generates a signature for the specified reference, which
	// can be a Porter bundle or an invocation image.
	Sign(ctx context.Context, ref string) error
	// Verify attempts to verify a signature for the specified
	// reference, which can be a Porter bundle or an invocation image.
	Verify(ctx context.Context, ref string) error
}
