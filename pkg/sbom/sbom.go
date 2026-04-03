package sbom

import (
	"context"
)

type SBOMGenerator interface {
	Close() error
	Connect(ctx context.Context) error

	Generate(ctx context.Context, bundleRef string, sbomPath string, insecureRegistry bool) error
}
