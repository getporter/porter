package manifest

import "get.porter.sh/porter/pkg/config"

// MakeCNABCompatible receives a Paramaeter Definition with possible porter specific parameters
// and converts those parameters to CNAB compatible versions.
// Returns true if values were replaced and false otherwise.
func MakeCNABCompatible(def *ParameterDefinition) bool {
	if v, ok := def.Type.(string); ok {
		if c, ok := config.PorterParamMap[v]; ok {
			def.Type = c.Type
			def.ContentEncoding = c.Encoding
			return ok
		}
	}

	return false
}
