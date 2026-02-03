package plugins

import (
	"context"
)

// SBOMGeneratorProtocol is the interface that SBOM generator plugins must implement.
// This defines the protocol used to communicate with SBOM generator plugins.
type SBOMGeneratorProtocol interface {
	Connect(ctx context.Context) error
	// Generate a SBOM for a given bundle reference
	Generate(ctx context.Context, bundleRef string, sbomPath string, insecureRegistry bool) error
}
