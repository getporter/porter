package runtime

import (
	"os"

	"get.porter.sh/porter/pkg/manifest"
	"github.com/pkg/errors"
)

// ReadDependencyOutputValue reads the dependency's output using the alias for the dependency from the
// specified output parameter source (name).
func (m *RuntimeManifest) ReadDependencyOutputValue(ref manifest.DependencyOutputReference) (string, error) {
	ps := manifest.GetParameterSourceForDependency(ref)
	psEnvVar := manifest.ParamToEnvVar(ps)
	output, ok := os.LookupEnv(psEnvVar)
	if !ok {
		err := errors.Errorf("bundle dependency %s output %s was not passed into the runtime", ref.Dependency, ref.Output)
		return "", err
	}

	return output, nil
}
