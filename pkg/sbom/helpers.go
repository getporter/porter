package sbom

import "get.porter.sh/porter/pkg/sbom/plugins/mock"

var _ SBOMGenerator = &TestSBOMGenerator{}

type TestSBOMGenerator struct {
	PluginAdapter

	sbomGenerator *mock.SBOMGenerator
}

func NewTestSBOMGeneratorProvider() TestSBOMGenerator {
	sbomGenerator := mock.NewSBOMGenerator()
	return TestSBOMGenerator{
		PluginAdapter: NewPluginAdapter(sbomGenerator),
		sbomGenerator: sbomGenerator,
	}
}
