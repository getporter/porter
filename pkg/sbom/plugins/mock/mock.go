package mock

import (
	"context"

	"get.porter.sh/porter/pkg/sbom/plugins"
	"get.porter.sh/porter/pkg/tracing"
)

var _ plugins.SBOMGeneratorProtocol = &SBOMGenerator{}

// SBOMGenerator implements an in-memory sbomGenerator for testing.
type SBOMGenerator struct {
}

func NewSBOMGenerator() *SBOMGenerator {
	s := &SBOMGenerator{}

	return s
}

func (s *SBOMGenerator) Generate(ctx context.Context, bundleRef string, sbomPath string, insecureRegistry bool) error {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	log.Infof("Generated SBOM signature for bundle %s", bundleRef)
	return nil
}

func (s *SBOMGenerator) Connect(ctx context.Context) error {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()
	return nil
}
