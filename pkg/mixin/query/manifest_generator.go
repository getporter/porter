package query

import (
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/pkg/errors"
)

// ManifestGenerator generates mixin input from the manifest contents associated with each mixin.
type ManifestGenerator struct {
	Manifest *manifest.Manifest
}

func NewManifestGenerator(m *manifest.Manifest) *ManifestGenerator {
	return &ManifestGenerator{
		Manifest: m,
	}
}

var _ MixinInputGenerator = &ManifestGenerator{}

func (g ManifestGenerator) ListMixins() []string {
	mixinNames := make([]string, len(g.Manifest.Mixins))
	for i, mixin := range g.Manifest.Mixins {
		mixinNames[i] = mixin.Name
	}
	return mixinNames
}

func (g ManifestGenerator) BuildInput(mixinName string) ([]byte, error) {
	input := g.buildInputForMixin(mixinName)
	inputB, err := yaml.Marshal(input)
	return inputB, errors.Wrapf(err, "could not marshal mixin build input for %s", mixinName)
}

func (g ManifestGenerator) buildInputForMixin(mixinName string) BuildInput {
	input := BuildInput{
		Actions: make(map[string]interface{}, 3),
	}

	for _, mixinDecl := range g.Manifest.Mixins {
		if mixinName == mixinDecl.Name {
			input.Config = mixinDecl.Config
		}
	}

	filterSteps := func(action string, steps manifest.Steps) {
		mixinSteps := manifest.Steps{}
		for _, step := range steps {
			if step.GetMixinName() != mixinName {
				continue
			}
			mixinSteps = append(mixinSteps, step)
		}
		input.Actions[action] = mixinSteps
	}
	filterSteps(cnab.ActionInstall, g.Manifest.Install)
	filterSteps(cnab.ActionUpgrade, g.Manifest.Upgrade)
	filterSteps(cnab.ActionUninstall, g.Manifest.Uninstall)

	for action, steps := range g.Manifest.CustomActions {
		filterSteps(action, steps)
	}

	return input
}
