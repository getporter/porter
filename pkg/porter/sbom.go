package porter

import (
	"get.porter.sh/porter/pkg/cnab"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/pkg/cataloger"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
	"github.com/syft/format/spdxjson"
)

func (p *Porter) GetSyftSBOM(ref cnab.OCIReference) error {
	detection, err := source.Detect(ref.String(), source.DefaultDetectConfig())
	src, err := detection.NewSource(source.DefaultDetectionSourceConfig())
	cataloger := cataloger.DefaultConfig()
	catalog, relationships, theDistro, err := syft.CatalogPackages(src, cataloger)
	if err != nil {
		return nil
	}

	sbom := &sbom.SBOM{
		Artifacts: sbom.Artifacts{
			Packages:          catalog,
			LinuxDistribution: theDistro,
		},
		Source:        src.Describe(),
		Relationships: relationships,
	}

	spdxjson.Encode(p.Out, sbom)

	return nil
}
