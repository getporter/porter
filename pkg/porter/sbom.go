package porter

import (
	"context"
	"errors"
	"path/filepath"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/format/spdxjson"
	"github.com/carolynvs/aferox"
)

type SBOMOptions struct {
	fileSystem aferox.Aferox
}

type SBOMGenerator interface {
	Generate(ctx context.Context, bundleRef cnab.BundleReference, sbomPath string) error
}

func NewSBOMGenerator(fileSystem aferox.Aferox) SBOMGenerator {
	return &SBOMOptions{
		fileSystem: fileSystem,
	}
}

func (s *SBOMOptions) Generate(ctx context.Context, bundleRef cnab.BundleReference, sbomPath string) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	log.Infof("Generating SBOM for bundle %s...", bundleRef.Reference.String())

	source, err := syft.GetSource(ctx, bundleRef.String(), syft.DefaultGetSourceConfig())
	if err != nil {
		return log.Errorf("failed to create source bundle %s: %w", bundleRef.Reference.String(), err)
	}
	sbom, err := syft.CreateSBOM(ctx, source, syft.DefaultCreateSBOMConfig().WithTool("Porter", pkg.Version))
	if err != nil {
		return log.Errorf("failed to create SBOM for bundle %s: %w", bundleRef.Reference.String(), err)
	}
	log.Infof("SBOM for bundle %s generated successfully", bundleRef.Reference.String())

	err = s.fileSystem.MkdirAll(filepath.Dir(sbomPath), 0o755)
	if err != nil {
		return log.Errorf("failed to create directory for SBOM %s: %w", sbomPath, err)
	}

	f, err := s.fileSystem.Create(sbomPath)
	if err != nil {
		return log.Errorf("failed to create SBOM file %s: %w", sbomPath, err)
	}
	defer func() {
		err = errors.Join(err, f.Close())
	}()

	encoder, err := spdxjson.NewFormatEncoderWithConfig(spdxjson.DefaultEncoderConfig())
	if err != nil {
		return log.Errorf("failed to create SBOM encoder: %w", err)
	}
	err = encoder.Encode(f, *sbom)
	if err != nil {
		return log.Errorf("failed to encode SBOM: %w", err)
	}

	log.Infof("SBOM for bundle %s written to %s", bundleRef.Reference.String(), sbomPath)
	return err
}
