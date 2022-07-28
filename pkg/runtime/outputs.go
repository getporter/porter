package runtime

import (
	"fmt"

	"get.porter.sh/porter/pkg/manifest"
)

// ReadDependencyOutputValue reads the dependency's output using the alias for the dependency from the
// specified output parameter source (name).
func (m *RuntimeManifest) ReadDependencyOutputValue(ref manifest.DependencyOutputReference) (string, error) {
	ps := manifest.GetParameterSourceForDependency(ref)
	psEnvVar := manifest.ParamToEnvVar(ps)
	output, ok := m.config.LookupEnv(psEnvVar)
	if !ok {
		err := fmt.Errorf("bundle dependency %s output %s was not passed into the runtime", ref.Dependency, ref.Output)
		return "", err
	}

	return output, nil
}
